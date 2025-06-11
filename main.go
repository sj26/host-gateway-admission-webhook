package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
	defaulter     = runtime.ObjectDefaulter(runtimeScheme)
)

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1.AddToScheme(runtimeScheme)
	_ = admissionv1.AddToScheme(runtimeScheme)
}

func HandleMutate(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "expected `application/json` content type", http.StatusUnsupportedMediaType)
		return
	}

	admissionReview := admissionv1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
		http.Error(w, fmt.Sprintf("invalid body: %v", err), http.StatusBadRequest)
		return
	}

	admissionReview.Response = &admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}

	if admissionReview.Request.Kind.Kind == "Pod" {
		var pod corev1.Pod
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
			http.Error(w, fmt.Sprintf("invalid object: %v", err), http.StatusBadRequest)
			return
		}

		var patches []map[string]any
		var ip string
		for i, alias := range pod.Spec.HostAliases {
			if alias.IP == "host-gateway" {
				if ip == "" {
					ips, err := net.LookupIP("host.docker.internal")
					if err != nil {
						http.Error(w, fmt.Sprintf("error resolving host.docker.internal: %v", err), http.StatusInternalServerError)
						return
					}
					if len(ips) == 0 {
						http.Error(w, "no address found for host.docker.internal", http.StatusInternalServerError)
						return
					}
					ip = ips[0].String()
				}

				patch := map[string]any{
					"op":    "replace",
					"path":  fmt.Sprintf("/spec/hostAliases/%d/ip", i),
					"value": ip,
				}
				patches = append(patches, patch)
			}
		}

		if len(patches) > 0 {
			log.Printf("[%s] Pod/%s: replacing `host-gateway` with `%s`", pod.Namespace, pod.Name, ip)

			patchType := admissionv1.PatchTypeJSONPatch
			admissionReview.Response.PatchType = &patchType

			patchBytes, err := json.Marshal(patches)
			if err != nil {
				http.Error(w, fmt.Sprintf("patch marshal error: %v", err), http.StatusInternalServerError)
				return
			}
			admissionReview.Response.Patch = patchBytes
		}
	}

	err := json.NewEncoder(w).Encode(admissionReview)
	if err != nil {
		http.Error(w, fmt.Sprintf("response encode error: %v", err), http.StatusInternalServerError)
	}
}

func main() {
	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "443"
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	// Yes this is not ideal, but it works
	certString := os.Getenv("CERT")
	keyString := os.Getenv("KEY")

	cert, err := tls.X509KeyPair([]byte(certString), []byte(keyString))
	if err != nil {
		log.Fatalf("Invalid certificate: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/mutate", HandleMutate)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	log.Fatal(srv.ListenAndServeTLS("", ""))
}
