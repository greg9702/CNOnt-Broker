import React from 'react';
import ReactJson from 'react-json-view';
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

    downloadFile = (blob, fileName) => {
        const link = document.createElement('a');
        link.href = URL.createObjectURL(blob);
        link.download = fileName;
        document.body.append(link);
        link.click();
        link.remove();
        setTimeout(() => URL.revokeObjectURL(link.href), 7000);
    }
    serializeClusterConfig = () => {

        fetch(this.state.serverUrl + "/api/v1/serialize-cluster-conf")
            .then((response) => {
                if (response.status === 200) {
                    const b = new Blob([response.blob()]);
                    this.downloadFile(b, "cluster_mapping.owl");
                    this.setState({message: "Obtained generated file"});
                } else if (response.status === 404) {
                    this.setState({message: "Generated file not found"});
                } else if (response.status === 409) {
                    this.setState({message: "Generating mapping error"});
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
            message: "",
            preview: null
        })
    }

    render() {
        return (
            <div id="deployment-manager">
                <div id="command-buttons">
                    <button
                        id="command-button"
                        onClick={this.state.preview === null ? this.fetchDeploymentPreview : this.closeDeploymentPreview}>
                        {this.state.preview === null ? "Show deployment preview" : "Close deployment preview"}
                    </button>

                    <button 
                        id="command-button"
                        onClick={this.sendCreateDeployment}>
                        Create deployment
                    </button>

                    <button 
                        id="command-button"
                        onClick={this.sendDeleteDeployment}>
                        Delete deployment
                    </button>
                    <button 
                        id="command-button"
                        onClick={this.serializeClusterConfig}>
                        Create mapping
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
