import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Navbar from './components/Navbar';
import ComponentList from './pages/ComponentList';
import DAG from './pages/DAG';
import StatusBuildInfo from './pages/status/BuildInfo';
import StatusFlags from './pages/status/Flags';
import StatusConfigFile from './pages/status/ConfigFile';

function App() {
  return (
    <BrowserRouter>
      <Navbar />
      <Routes>
        <Route path="/" element={<ComponentList />} />
        <Route path="/dag" element={<DAG />} />
        <Route path="/status/build-info" element={<StatusBuildInfo />} />
        <Route path="/status/flags" element={<StatusFlags />} />
        <Route path="/status/config-file" element={<StatusConfigFile />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
