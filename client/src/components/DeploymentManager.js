import React from 'react';
import './DeploymentManager.css';

export default class EchoSender extends React.Component {
  
    state = {
      serverUrl: "http://localhost:8080",
      message: null
    }

    sendCreateDeployment = () => {
      fetch(this.state.serverUrl + "/api/v1/create-deployment")
      .then((response) => {
        if (response.status == 201) {
          this.setState({ message: "Created sucessfully" });
        } else {
          throw Error(response.statusText)
        }
      }).catch(error => {
        this.setState({ message: "Creating " + error });
      })
    }

    sendDeleteDeployment = () => {
      fetch(this.state.serverUrl + "/api/v1/delete-deployment")
      .then((response) => {
        if (response.status == 204) {
          this.setState({ message: "Deployment deleted sucessfully" });
        } else if (response.status == 404) {
          this.setState({ message: "Deployment do not exists" });
        } else {
          throw Error(response.statusText)
        }
      }).catch(error => {
        this.setState({ message: "Deleting " + error });
      })
    }

    render() {
      return (
        <div id="deployment-manager">
          <button
            onClick={this.sendCreateDeployment}> 
            Create deployment
          </button>

          <button 
            onClick={this.sendDeleteDeployment}> 
            Delete deployment
          </button>
          
          <div>
            {this.state.message === null ? "" : this.state.message }
          </div>
        </div>
      );
    }
  }
  