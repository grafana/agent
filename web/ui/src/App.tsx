import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Navbar from './components/Navbar';
import ComponentList from './pages/ComponentList';
import DAG from './pages/DAG';

function App() {
  return (
    <BrowserRouter>
      <Navbar />
      <Routes>
        <Route path="/" element={<ComponentList />} />
        <Route path="/dag" element={<DAG />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
