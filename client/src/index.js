import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';

function sleep (time) {
  return new Promise((resolve) => setTimeout(resolve, time));
}

class MessageReceiver extends React.Component {
  
  state = {
    serverUrl: "http://localhost:8080",
    loadingState: false,
    message: "empty",
    number: null
  }

  getHello = () => {
    this.setState({ loadingState: true });

    fetch(this.state.serverUrl + "/api/v1/hello")
      .then(res => res.json())
      .then(
        (result) => {
          this.setState({ message: result["message"] });
        },
        (error) => {
          this.setState({ message: "error fetching message" });
        }
      )

    sleep(1000).then(() => {
      // Do something after the sleep!
      this.setState({ loadingState: false });
    });
  }

  sendEcho = (numberBox) => {
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
      <div>
        <div id="hello-receiver">
          <button onClick={this.getHello}>
          Get Hello
          </button><br/>
        
          {this.state.loadingState ? 
            <div>Loading...</div> :
            <div>Message: {this.state.message}</div>}
        </div>
        
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
          {this.state.number == "error" ? 'Error!' :
            this.state.number == null ? "No number" : 
              'Received number ' + this.state.number }
          </div>
        </div>
        
      </div>
    );
  }
}

ReactDOM.render(<MessageReceiver />, document.getElementById('root'));