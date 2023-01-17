local pipelines = import './pipelines.jsonnet';

(import 'pipelines/build_images.jsonnet') +
(import 'pipelines/test.jsonnet') +
(import 'pipelines/check_containers.jsonnet') +
(import 'pipelines/publish.jsonnet') +
(import 'secrets.jsonnet')
