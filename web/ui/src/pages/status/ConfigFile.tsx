import { faGear } from '@fortawesome/free-solid-svg-icons';
import Page from '../../components/Page';

function ConfigFile() {
  return <Page name="Config file" desc="Last successfully read unevaluated config file" icon={faGear} />;
}

export default ConfigFile;
