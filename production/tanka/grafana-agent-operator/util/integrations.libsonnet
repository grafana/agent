local gen = import 'agent-operator-gen/main.libsonnet';
local int = gen.monitoring.v1alpha1.integration;

{
    withPVC(name):: {
        spec+: {
            volumeMounts: [
                int.spec.volumeMounts.withName(name) +
                int.spec.volumeMounts.withMountPath('/etc/eventhandler')
            ],
            volumes: [
                int.spec.volumes.withName(name) +
                int.spec.volumes.persistentVolumeClaim.withClaimName(name)
            ]
        }
    }
}
