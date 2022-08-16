import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Navbar from './components/Navbar';
import PageComponentList from './pages/PageComponentList';
import DAG from './pages/DAG';
import StatusBuildInfo from './pages/status/BuildInfo';
import StatusFlags from './pages/status/Flags';
import StatusConfigFile from './pages/status/ConfigFile';
import styles from './App.module.css';
import { ComponentDetailPage } from './pages/ComponentDetailPage';

function App() {
  return (
    <div className={styles.app}>
      <BrowserRouter>
        <Navbar />
        <main>
          <Routes>
            <Route path="/" element={<PageComponentList />} />
            <Route path="/dag" element={<DAG />} />
            <Route path="/status/build-info" element={<StatusBuildInfo />} />
            <Route path="/status/flags" element={<StatusFlags />} />
            <Route path="/status/config" element={<StatusConfigFile />} />
            <Route path="/component/:component" element={<ComponentDetailPage />} />
          </Routes>
        </main>
      </BrowserRouter>
    </div>
  );
}

export default App;
