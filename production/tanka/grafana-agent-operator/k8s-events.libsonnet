local gen = import 'agent-operator-gen/main.libsonnet';
local int = gen.monitoring.v1alpha1.integration;
local k = import 'ksonnet-util/kausal.libsonnet';
local pvc = k.core.v1.persistentVolumeClaim;
  
{
    new(name, namespace, integrationLabels, logsInstanceName, logsInstanceNamespace):: {    
        resource:
            int.new('agent-eventhandler') +
            int.metadata.withNamespace(namespace) +
            int.metadata.withLabels(integrationLabels) +
            int.spec.type.withUnique(true) +
            int.spec.withConfig({
                logs_instance: logsInstanceNamespace + '/' + logsInstanceName,
                cache_path: '/etc/eventhandler/eventhandler.cache',
            })
    },

    withPVC(name, namespace):: {
        pvc:
            pvc.new(name) +
            pvc.mixin.metadata.withNamespace(namespace) +
            pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
            pvc.mixin.spec.resources.withRequests({ storage: '1Gi' }),
        
        resource+: {
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
}
