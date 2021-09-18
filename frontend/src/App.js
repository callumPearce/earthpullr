import React from 'react';
import logo from './logo.png';
import './App.css';
import ImagesRetriever from './components/ImagesRetriever';

function App() {
  return (
    <div id="app" className="App">
      <header className="App-header">
        <ImagesRetriever />
      </header>
    </div>
  );
}

export default App;
