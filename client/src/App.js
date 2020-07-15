import React from 'react';
import './index.css';
import MessageReceiver from './components/MessageReceiver' 
import EchoSender from './components/EchoSender' 
import DeploymentManager from './components/DeploymentManager'
import './App.css';

export default class App extends React.Component {
    
    render() {
      return (
        <DeploymentManager />
      );
    }
}
  