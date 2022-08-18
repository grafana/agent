import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Navbar from './features/layout/Navbar';
import PageComponentList from './pages/PageComponentList';
import DAG from './pages/DAG';
import StatusBuildInfo from './pages/status/BuildInfo';
import StatusFlags from './pages/status/Flags';
import StatusConfigFile from './pages/status/ConfigFile';
import styles from './App.module.css';
import { ComponentDetailPage } from './pages/ComponentDetailPage';
import { PathPrefixContext } from './contexts/PathPrefixContext';

function App() {
  // Production builds set REACT_APP_BASE_URL to a template string which Go
  // code will replace with the value of --server.http.ui-path-prefix when
  // serving this page.
  //
  // When developing with `yarn run watch`, this field is unset and is set to
  // the root path.
  const baseName = process.env.REACT_APP_BASE_URL || '/';

  return (
    <PathPrefixContext.Provider value={baseName}>
      <div className={styles.app}>
        <BrowserRouter basename={baseName}>
          <Navbar />
          <main>
            <Routes>
              <Route path="/" element={<PageComponentList />} />
              <Route path="/components" element={<PageComponentList />} />
              <Route path="/component/:component" element={<ComponentDetailPage />} />
              <Route path="/dag" element={<DAG />} />
              <Route path="/status/build-info" element={<StatusBuildInfo />} />
              <Route path="/status/config" element={<StatusConfigFile />} />
              <Route path="/status/flags" element={<StatusFlags />} />
            </Routes>
          </main>
        </BrowserRouter>
      </div>
    </PathPrefixContext.Provider>
  );
}

export default App;
