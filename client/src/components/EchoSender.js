import React from 'react';
import './EchoSender.css';

export default class EchoSender extends React.Component {
  
    state = {
      serverUrl: "http://localhost:8080",
      number: null
    }
  
    sendEcho = () => {
      fetch(this.state.serverUrl + "/api/v1/hello/" + this.refs.topic.value)
      .then(res => res.json())
      .then(
        (result) => {
          this.setState({ number: result["number"] });
        },
        (error) => {
          this.setState({ number: "error" });
        }
      )
    }

    render() {
      return (
        <div id="echo-sender">
          <input 
            ref="topic"
            type="text"
            name="numberBox"
            placeholder="Enter number here..."/>

          <button 
            value="Send"
            onClick={this.sendEcho}> 
            Send
          </button>
          
          <div>
            {this.state.number === "error" ? 'Error!' :
              this.state.number == null ? "No number" : 
                'Received number ' + this.state.number }
          </div>
        </div>
      );
    }
  }
  