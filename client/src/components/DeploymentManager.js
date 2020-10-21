import React from 'react';
import ReactJson from 'react-json-view'
import './DeploymentManager.css';

export default class DeploymentManager extends React.Component {

    state = {
        serverUrl: "http://localhost:8080",
        message: null,
        preview: null
    }

    sendCreateDeployment = () => {
        fetch(this.state.serverUrl + "/api/v1/create-deployment")
            .then((response) => {
                if (response.status === 201) {
                    this.setState({message: "Created sucessfully"});
                } else {
                    throw Error(response.statusText)
                }
            }).catch(error => {
            this.setState({message: "Creating " + error});
        })
    }

    sendDeleteDeployment = () => {
        fetch(this.state.serverUrl + "/api/v1/delete-deployment")
            .then((response) => {
                if (response.status === 204) {
                    this.setState({message: "Deployment deleted sucessfully"});
                } else if (response.status === 404) {
                    this.setState({message: "Deployment do not exists"});
                } else {
                    throw Error(response.statusText)
                }
            }).catch(error => {
            this.setState({message: "Deleting " + error});
        })
    }

    serializeClusterConfig = () => {
        fetch(this.state.serverUrl + "/api/v1/serialize-cluster-conf")
            .then((response) => {
                if (response.status === 200) {
                        this.setState({message: "Received pods list (see core logs)"});
                } else if (response.status === 404) {
                    this.setState({message: "Error obtaining pods list"});
                } else {
                    throw Error(response.statusText)
                }
            }).catch(error => {
            this.setState({message: "Serializer error" + error});
        })
    }

    fetchDeploymentPreview = () => {
        fetch(this.state.serverUrl + "/api/v1/preview-deployment")
            .then((response) => {
                if (response.status === 200) {
                    response.json().then(data => {
                        console.log(data["deployment"])
                        this.setState({
                            message: "Deployment preview available",
                            preview: data["deployment"]
                        })
                    });
                } else if (response.status === 404) {
                    this.setState({message: "Deployment do not exists"});
                } else {
                    throw Error(response.statusText)
                }
            }).catch(error => {
            this.setState({message: "Deployment preview error: " + error});
        })
    }

    closeDeploymentPreview = () => {
        this.setState({
            message: "Deployment preview hidden",
            preview: null
        })
    }

    render() {
        return (
            <div id="deployment-manager">
                <div id="deployment-creator">
                </div>
                <div id="command-buttons">
                    <button
                        onClick={this.state.preview === null ? this.fetchDeploymentPreview : this.closeDeploymentPreview}>
                        {this.state.preview === null ? "Show deployment preview" : "Close deployment preview"}
                    </button>

                    <button
                        onClick={this.sendCreateDeployment}>
                        Create deployment
                    </button>

                    <button
                        onClick={this.sendDeleteDeployment}>
                        Delete deployment
                    </button>
                    <button
                        onClick={this.serializeClusterConfig}>
                        Serialize (WIP)
                    </button>
                </div>
                <div id="status-logger">
                    {this.state.message === null ? "" : this.state.message}
                </div>
                <div id="deployment-preview" style={this.state.preview === null ? {opacity: '0%'} : {opacity: '100%'}}>
                    {this.state.preview === null ? "" :
                        <ReactJson src={this.state.preview}
                                   name="Deployment preview"
                                   theme="pop"
                                   iconStyle="square"
                                   indentWidth={10}
                                   displayObjectSize={false}
                                   displayDataTypes={false}/>}
                </div>
            </div>
        );
    }
}
