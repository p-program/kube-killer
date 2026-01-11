package killer

import (
	"context"
	"math/rand"
	"time"

	"github.com/p-program/kube-killer/core"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/pager"
)

type SecretKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	half         bool
	namespace    string
}

func NewSecretKiller(namespace string) (*SecretKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := SecretKiller{
		namespace: namespace,
		client:    clientset,
	}
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *SecretKiller) DryRun() *SecretKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *SecretKiller) BlackHand() *SecretKiller {
	k.mafia = true
	return k
}

func (k *SecretKiller) SetHalf() *SecretKiller {
	k.half = true
	return k
}

func (k *SecretKiller) getAllSecretInCurrentNamespace() ([]*v1.Secret, error) {
	p := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		list, err := k.client.CoreV1().Secrets(k.namespace).List(context.TODO(), opts)
		if err != nil {
			return nil, errors.Wrap(err, "cannot retrieve secrets")
		}
		return list, nil
	}))
	p.PageSize = 500
	ctx := context.Background()
	secrets := []*v1.Secret{}
	err := p.EachListItem(ctx, metav1.ListOptions{
		FieldSelector: "type=Opaque",
	}, func(obj runtime.Object) error {
		secret, ok := obj.(*v1.Secret)
		if !ok {
			return errors.Errorf("this is not a secret: %#v", obj)
		}
		secrets = append(secrets, secret)
		return nil
	})
	if err != nil {
		return []*v1.Secret{}, errors.Wrap(err, "cannot iterate secrets")
	}
	return secrets, nil
}

func (k *SecretKiller) Kill() error {
	secrets, err := k.getAllSecretInCurrentNamespace()
	if err != nil {
		return err
	}
	
	if k.mafia {
		if k.half {
			return k.KillHalfSecrets(secrets)
		}
		return k.KillAllSecrets(secrets)
	}
	
	podKiller, err := NewPodKiller(k.namespace)
	if err != nil {
		return err
	}
	pods, err := podKiller.getAllPodsInCurrentNamespace()
	if err != nil {
		return err
	}
	
	// Get ServiceAccounts that use secrets
	serviceAccounts, err := k.getAllServiceAccountsInCurrentNamespace()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to retrieve ServiceAccounts, continuing without them")
	}
	
	unusedSecret, err := k.detectUnusedSecrets(pods, secrets, serviceAccounts)
	if err != nil {
		return err
	}
	for _, secret := range unusedSecret {
		log.Info().Msgf("Deleting unused secret %s in namespace %s", secret.Name, k.namespace)
		err = k.client.CoreV1().Secrets(k.namespace).Delete(context.TODO(), secret.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return err
}

func (k *SecretKiller) KillAllSecrets(secrets []*v1.Secret) error {
	for _, secret := range secrets {
		log.Info().Msgf("Deleting secret %s in namespace %s", secret.Name, k.namespace)
		err := k.client.CoreV1().Secrets(k.namespace).Delete(context.TODO(), secret.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *SecretKiller) KillHalfSecrets(secrets []*v1.Secret) error {
	if len(secrets) == 0 {
		log.Info().Msg("No secrets to kill")
		return nil
	}
	
	// Randomly shuffle the secrets list
	secretList := make([]*v1.Secret, len(secrets))
	copy(secretList, secrets)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(secretList), func(i, j int) {
		secretList[i], secretList[j] = secretList[j], secretList[i]
	})
	
	// Calculate how many secrets to kill (half, rounded down)
	secretsToKill := len(secretList) / 2
	if secretsToKill == 0 {
		secretsToKill = 1 // At least kill one secret if there's only one
	}
	
	log.Info().Msgf("Killing %d out of %d secrets", secretsToKill, len(secretList))
	for i := 0; i < secretsToKill; i++ {
		secret := secretList[i]
		log.Info().Msgf("Deleting secret %s in namespace %s", secret.Name, k.namespace)
		err := k.client.CoreV1().Secrets(k.namespace).Delete(context.TODO(), secret.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *SecretKiller) getAllServiceAccountsInCurrentNamespace() ([]v1.ServiceAccount, error) {
	saList, err := k.client.CoreV1().ServiceAccounts(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return saList.Items, nil
}

func (k *SecretKiller) detectUnusedSecrets(pods []*v1.Pod, secrets []*v1.Secret, serviceAccounts []v1.ServiceAccount) ([]*v1.Secret, error) {
	usedSecretNames := map[string]bool{}
	
	// Check secrets used by Pods
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			for _, envFrom := range container.EnvFrom {
				if envFrom.SecretRef != nil {
					usedSecretNames[envFrom.SecretRef.Name] = true
				}
			}
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
					usedSecretNames[env.ValueFrom.SecretKeyRef.Name] = true
				}
			}
		}
		for _, volume := range pod.Spec.Volumes {
			if volume.Secret != nil {
				usedSecretNames[volume.Secret.SecretName] = true
			}
			if volume.Projected != nil {
				for _, source := range volume.Projected.Sources {
					if source.Secret != nil {
						usedSecretNames[source.Secret.Name] = true
					}
				}
			}
		}
	}
	
	// Check secrets used by ServiceAccounts
	for _, sa := range serviceAccounts {
		for _, secretRef := range sa.Secrets {
			usedSecretNames[secretRef.Name] = true
		}
		for _, imagePullSecret := range sa.ImagePullSecrets {
			usedSecretNames[imagePullSecret.Name] = true
		}
	}
	
	unused := []*v1.Secret{}
	for _, secret := range secrets {
		if !usedSecretNames[secret.Name] {
			unused = append(unused, secret)
		}
	}

	return unused, nil
}
