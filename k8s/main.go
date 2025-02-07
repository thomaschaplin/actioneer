package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type RequestBody struct {
	Namespace   string `json:"namespace"`
	EndpointURL string `json:"endpointURL"`
}

func main() {
	http.HandleFunc("/webhook", receiveHandler)
	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func receiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var reqBody RequestBody
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &reqBody); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received namespace: %s, endpointURL: %s\n", reqBody.Namespace, reqBody.EndpointURL)

	http.Get(reqBody.EndpointURL)

	// Create Kubernetes Pod with the extracted namespace and endpoint URL
	if err := createCurlPod("actioneer-operator", "www.google.com"); err != nil {
		http.Error(w, "Failed to create Kubernetes pod", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Kubernetes pod created successfully"))
}

func createCurlPod(namespace, endpointURL string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil // return nil if not running in Kubernetes
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %v", err)
	}

	// Generate a unique pod name using the namespace and UUID
	// Limit the pod name length to 63 characters to avoid DNS issues
	podName := fmt.Sprintf("%s-curl-pod-%s", namespace, uuid.New().String())

	// Ensure the pod name does not exceed 63 characters
	if len(podName) > 63 {
		podName = podName[:63] // Truncate to 63 characters
	}

	// Make sure the pod name only contains valid DNS label characters
	podName = strings.ToLower(podName)
	podName = strings.ReplaceAll(podName, "_", "-")

	// Proceed with creating the pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "curl-container",
					Image: "runner:latest",
					// Args:            []string{"-X", "GET", endpointURL},
					ImagePullPolicy: corev1.PullNever,
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8080,
							Name:          "http",
							Protocol:      corev1.ProtocolTCP,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "repo-volume",
							MountPath: "/repo", // Mount the volume inside the container at /repo
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
			Volumes: []corev1.Volume{
				{
					Name: "repo-volume",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{}, // Creates an ephemeral volume
					},
				},
			},
		},
	}

	_, err = clientset.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pod: %v", err)
	}

	return nil
}

// docker build --platform=linux/amd64 -t github-actions-poc-service .
// kind load docker-image github-actions-poc-service:latest
// smee -u https://smee.io/1eOP0ZmjbtoBY9XS -p 8080 -t http://127.0.0.1:8080/webhook
// docker tag github-actions-poc-service:latest
// docker push
