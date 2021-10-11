import React from 'react';
import logo from './logo.png';
import './App.css';
import ImagesRetriever from './components/ImagesRetriever';

function App() {
  return (
    <div id="app" className="App">
      <header className="App-header">
          <meta name="viewport" content="initial-scale=1, width=device-width" />
        <ImagesRetriever />
      </header>
    </div>
  );
}

export default App;
