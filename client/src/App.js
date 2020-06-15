import React from 'react';
import './index.css';
import MessageReceiver from './components/MessageReceiver' 
import EchoSender from './components/EchoSender' 
import DeploymentManager from './components/DeploymentManager' 

export default class App extends React.Component {
    
    render() {
      return (
        <div>
          <DeploymentManager />
        </div>
      );
    }
}
  