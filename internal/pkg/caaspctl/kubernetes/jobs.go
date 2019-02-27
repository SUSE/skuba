package kubernetes

import (
	"errors"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateJob(name string, spec batchv1.JobSpec) (*batchv1.Job, error) {
	return GetAdminClientSet().BatchV1().Jobs(metav1.NamespaceSystem).Create(&batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
		},
		Spec: spec,
	})
}

func DeleteJob(name string) error {
	return GetAdminClientSet().BatchV1().Jobs(metav1.NamespaceSystem).Delete(name, &metav1.DeleteOptions{})
}

func CreateAndWaitForJob(name string, spec batchv1.JobSpec) error {
	_, err := CreateJob(name, spec)
	defer DeleteJob(name)
	if err != nil {
		return err
	}
	for i := 0; i < 60; i++ {
		job, err := GetAdminClientSet().BatchV1().Jobs(metav1.NamespaceSystem).Get(name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("failed to get status for job %s, continuing...\n", name)
		} else {
			if job.Status.Active > 0 {
				fmt.Printf("job %s is active, waiting...\n", name)
			} else {
				if job.Status.Succeeded > 0 {
					fmt.Printf("job %s executed successfully\n", name)
					return nil
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
	return errors.New(fmt.Sprintf("failed waiting for job %s", name))
}
