import React from 'react';
import './index.css';
import MessageReceiver from './components/MessageReceiver' 
import EchoSender from './components/EchoSender' 

export default class App extends React.Component {
    
    render() {
      return (
        <div>
          <MessageReceiver />
          <EchoSender />
        </div>
      );
    }
}
  