import React from 'react';
import { BrowserRouter, Route, Routes } from 'react-router-dom';

import Navbar from './features/layout/Navbar';
import ComponentDetailPage from './pages/ComponentDetailPage';
import Graph from './pages/Graph';
import PageComponentList from './pages/PageComponentList';

interface Props {
  basePath: string;
}

const Router = ({ basePath }: Props) => {
  return (
    <BrowserRouter basename={basePath}>
      <Navbar />
      <main>
        <Routes>
          <Route path="/" element={<PageComponentList />} />
          <Route path="/component/*" element={<ComponentDetailPage />} />
          <Route path="/graph" element={<Graph />} />
        </Routes>
      </main>
    </BrowserRouter>
  );
};

export default Router;
