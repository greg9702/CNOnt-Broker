import React from 'react';
import './index.css';
import DeploymentManager from './components/DeploymentManager'
import './App.css';

export default class App extends React.Component {
    
    render() {
      return (
        <div id="app">
          <div id="header" className="eye-catching">
              CNOnt Broker App
          </div>
          <DeploymentManager />
        </div>
      );
    }
}
