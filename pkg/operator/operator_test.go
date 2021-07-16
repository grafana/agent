package operator

import (
	"fmt"
	"testing"

	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestOperator(t *testing.T) {
	instList := &gragent.GrafanaAgentList{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "GrafanaAgentList",
			APIVersion: gragent.SchemeGroupVersion.Identifier(),
		},
		ListMeta: meta_v1.ListMeta{},
		Items: []*gragent.GrafanaAgent{
			{ObjectMeta: meta_v1.ObjectMeta{Name: "a"}},
			{ObjectMeta: meta_v1.ObjectMeta{Name: "b"}},
		},
	}
	_ = instList

	meta.EachListItem(instList, func(o runtime.Object) error {
		fmt.Printf("%#v\n", o)
		return nil
	})

	/*
		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(instList)
		require.NoError(t, err)
		us := unstructured.Unstructured{Object: data}

		_ = us.EachListItem(func(o runtime.Object) error {
			us := o.(*unstructured.Unstructured)

			var a gragent.GrafanaAgent
			runtime.DefaultUnstructuredConverter.FromUnstructured(us.Object, &a)
			fmt.Printf("%#v\n", a)
			return nil
		})
	*/

}
