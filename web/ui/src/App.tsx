import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Navbar from './features/layout/Navbar';
import PageComponentList from './pages/PageComponentList';
import Graph from './pages/Graph';
import styles from './App.module.css';
import { ComponentDetailPage } from './pages/ComponentDetailPage';
import { PathPrefixContext } from './contexts/PathPrefixContext';

/**
 * getBasePath retrieves the base path of the application by looking at the
 * <base> HTML element in the HTML header. If there is no <base> element or the
 * <base> element is empty, getBaseURL returns "/".
 */
function getBasePath(): string {
  const elements = document.getElementsByTagName('base');
  if (elements.length !== 1) {
    return '/';
  }

  // elements[0].href will be a full URL, but we just want to extract the path
  // portion.
  return new URL(elements[0].href).pathname;
}

function App() {
  const basePath = getBasePath();

  return (
    <PathPrefixContext.Provider value={basePath}>
      <div className={styles.app}>
        <BrowserRouter basename={basePath}>
          <Navbar />
          <main>
            <Routes>
              <Route path="/" element={<PageComponentList />} />
              <Route path="/components" element={<PageComponentList />} />
              <Route path="/component/:id" element={<ComponentDetailPage />} />
              <Route path="/graph" element={<Graph />} />
            </Routes>
          </main>
        </BrowserRouter>
      </div>
    </PathPrefixContext.Provider>
  );
}

export default App;
