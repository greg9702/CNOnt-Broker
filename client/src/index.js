import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';

function sleep (time) {
  return new Promise((resolve) => setTimeout(resolve, time));
}

class MessageReceiver extends React.Component {
  state = {
    loadingState: false,
    message: "empty",

  }

  getHello = () => {
    this.setState({ loadingState: true });

    fetch("http://localhost:8080/api/v1/hello")
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

  render() {
    return (
        <div>
        <button onClick={this.getHello}>
         Get Hello
        </button><br/>
       
        {this.state.loadingState ? 
          <div>Loading...</div> :
          <div>Message: {this.state.message}</div>}
      </div>
    );
  }
}

ReactDOM.render(<MessageReceiver />, document.getElementById('root'));