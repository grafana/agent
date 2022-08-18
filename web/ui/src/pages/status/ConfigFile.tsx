import { faGear } from '@fortawesome/free-solid-svg-icons';
import Page from '../../features/layout/Page';

function ConfigFile() {
  return <Page name="Config file" desc="Last successfully read unevaluated config file" icon={faGear} />;
}

export default ConfigFile;
