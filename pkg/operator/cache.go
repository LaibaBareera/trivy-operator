package operator

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

const lastAppliedConfigurationAnnotation = "kubectl.kubernetes.io/last-applied-configuration"

// CacheTransform returns the transform applied to every object entering the
// operator's shared informer cache. It strips managed fields, the
// last-applied-configuration annotation and the contents of every ConfigMap,
// all of which keep the cache small.
//
// Stripping ConfigMap contents is safe because ConfigMap is listed in the
// manager's client.CacheOptions.DisableFor (see Start), so every ConfigMap read
// through the manager client - including the one config-audit hands to Rego -
// goes to the API server and returns a full object.
func CacheTransform() toolscache.TransformFunc {
	stripManagedFields := cache.TransformStripManagedFields()

	return func(obj any) (any, error) {
		obj, err := stripManagedFields(obj)
		if err != nil {
			return obj, err
		}

		if metaObj, ok := obj.(metav1.ObjectMetaAccessor); ok {
			annotations := metaObj.GetObjectMeta().GetAnnotations()
			if annotations != nil {
				delete(annotations, lastAppliedConfigurationAnnotation)
				metaObj.GetObjectMeta().SetAnnotations(annotations)
			}
		}

		if cm, ok := obj.(*corev1.ConfigMap); ok {
			cm.Data = nil
			cm.BinaryData = nil
		}

		return obj, nil
	}
}
