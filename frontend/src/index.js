import React from 'react';
import ReactDOM from 'react-dom';
import 'core-js/stable';
import './index.css';
import App from './App';
import * as serviceWorker from './serviceWorker';

import * as Wails from '@wailsapp/runtime';
import CssBaseline from '@mui/material/CssBaseline';
import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';

Wails.Init(() => {
  ReactDOM.render(
    <React.StrictMode>
        <CssBaseline/>
      <App />
    </React.StrictMode>,
    document.getElementById("app")
  );
});

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
