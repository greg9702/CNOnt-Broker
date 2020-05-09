import React from 'react';

export default class MessageReceiver extends React.Component {
  
    state = {
      serverUrl: "http://localhost:8080",
      loadingState: false,
      message: "empty",
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
      this.setState({ loadingState: false });
    }
  
    render() {
      return (
        <div>
          <div id="hello-receiver">
            <button onClick={this.getHello}>
            Get Hello
            </button>
            <br/>
            {this.state.loadingState ? 
              <div>Loading...</div> :
              <div>Message: {this.state.message}</div>}
          </div>
        </div>
      );
    }
  }
  