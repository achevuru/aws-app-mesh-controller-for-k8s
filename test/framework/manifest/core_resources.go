package manifest

import (
	"go.uber.org/zap"

	//	"encoding/pem"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
	"bufio"

	"github.com/aws/aws-app-mesh-controller-for-k8s/test/framework"
)

type ManifestBuilder struct {
	Namespace            string
	ServiceDiscoveryType ServiceDiscoveryType

	// required when serviceDiscoveryType == CloudMapServiceDiscovery
	CloudMapNamespace string
}

func (b *ManifestBuilder) BuildDeployment(instanceName string, replicas int32, appImage string, containerPort int32,
	env []corev1.EnvVar, annotations map[string]string) *appsv1.Deployment {
	labels := b.buildNodeSelectors(instanceName)
	dpName := b.buildNodeName(instanceName)
	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: b.Namespace,
			Name:      dpName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: appImage,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: containerPort,
								},
							},
							Env: env,
						},
					},
				},
			},
		},
	}
	return dp
}

func (b *ManifestBuilder) BuildServiceWithSelector(instanceName string, containerPort int32, targetPort int) *corev1.Service {
	labels := b.buildNodeSelectors(instanceName)
	svcName := b.buildNodeName(instanceName)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: b.Namespace,
			Name:      svcName,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       containerPort,
					TargetPort: intstr.FromInt(targetPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func (b *ManifestBuilder) BuildServiceWithoutSelector(instanceName string, containerPort int32, targetPort int) *corev1.Service {
	svcName := b.buildServiceName(instanceName)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: b.Namespace,
			Name:      svcName,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:       containerPort,
					TargetPort: intstr.FromInt(targetPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func (b *ManifestBuilder) BuildK8SSecretsFromPemFile(pemFileBasePath string, tlsFiles []string,
	secretName string, f *framework.Framework) *corev1.Secret {
	certMap := make(map[string][]byte, len(tlsFiles))
	for _, tlsFile := range tlsFiles {
		pemFilePath := pemFileBasePath + tlsFile
		privateKeyFile, err := os.Open(pemFilePath)
		if err != nil {
			f.Logger.Error("PEM File Open error: ", zap.Error(err))
			return nil
		}

		pemFileInfo, _ := privateKeyFile.Stat()
		pemFileSize := pemFileInfo.Size()
		pemBytes := make([]byte, pemFileSize)

		buffer := bufio.NewReader(privateKeyFile)
		_, err = buffer.Read(pemBytes)

		certMap[tlsFile] = pemBytes
	}


	secret := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Namespace: "tls-e2e",
		},
		Immutable:  nil,
		Data:       certMap,
		StringData: nil,
		Type:       corev1.SecretTypeOpaque,
	}

	return secret
}

func (b *ManifestBuilder) buildNodeSelectors(instanceName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "timeout-app",
		"app.kubernetes.io/instance": instanceName,
	}
}

func (b *ManifestBuilder) buildNodeName(instanceName string) string {
	return instanceName
}

func (b *ManifestBuilder) buildServiceName(instanceName string) string {
	return instanceName
}
