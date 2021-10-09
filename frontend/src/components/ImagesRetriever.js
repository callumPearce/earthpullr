import React from 'react';
import * as Wails from '@wailsapp/runtime'
import fs from 'fs'

const MAX_RES = 7680
const MAX_BACKGROUND_COUNT = 50

const imageDimensionValidation = dim => {
	if (!dim) {
		return "Both background width and height must be specified"
	}
	if (dim < 1 || dim > MAX_RES) {
		return "Maximum size of backgrounds that can be retrieved " + MAX_RES.toString() + "px (8k) in both dimensions";
	}
	return null;
}

const backgroundCountValidation = count => {
	console.log(count);
	if (!count) {
		return "The number of backgrounds you wish to retrieve must be specified"
	}
	if (count > MAX_BACKGROUND_COUNT) {
		return "At most " + MAX_BACKGROUND_COUNT.toString() + " can be retrieved at one time"
	}
	if (count < 1) {
		return "At least 1 background must be retrieved"
	}
	return null;
}

const downloadPathValidation = path => {
	if (!path) {
		return "The directory path to which to save the backgrounds to must be specified"
	}
	return null;
}

class ImagesRetriever extends React.Component {
	constructor(props, context) {
		super(props);

		this.state = {
			form: {
				backgroundsCount: {value: 0, errMsg: "", valid: false, validator: backgroundCountValidation},
				imagesWidth: {value: 0, errMsg: "", valid: false, validator: imageDimensionValidation},
				imagesHeight: {value: 0, errMsg: "", valid: false, validator: imageDimensionValidation},
				downloadPath: {value: "", errMsg: "", valid: false, validator: downloadPathValidation},
			},
			imagesSaved: 0,
			responseMsg: ""
		};

		this.handleSubmit = this.handleSubmit.bind(this);
		this.handleUserInput = this.handleUserInput.bind(this);
	}

	imagesSaved () {
		Wails.Events.On("image_saved", imagesSaved => {
			this.setState({ imagesSaved: this.state.imagesSaved + 1 });
		});
	}

	handleSubmit(event) {
		let valid = true;
		for (const [name, value] of Object.entries(this.state.form)) {
			valid = this.validateField(name, value.value) && valid
		}
		// console.log(this.state);
		if (valid) {
			const request = {
				BackgroundsCount: parseInt(this.state.form.backgroundsCount.value),
				Width: parseInt(this.state.form.imagesWidth.value),
				Height: parseInt(this.state.form.imagesHeight.value),
				DownloadPath: this.state.form.downloadPath.value
			}
			window.backend.BackgroundRetriever.GetBackgrounds(request).then(result =>
				this.setState({responseMsg: result})
			).catch(err =>
				this.setState({responseMsg: err})
			)
		}
		else {
			this.setState({responseMsg: "Form is invalid"})
		}
		event.preventDefault()
	}

	validateField(name, value) {
		let newState = this.state;
		const errMsg = newState.form[name].validator(value);
		if (errMsg != null) {
			newState.form[name].valid = false;
		}
		else {
			newState.form[name].valid = true;
		}
		newState.form[name].errMsg = errMsg;
		console.log(errMsg)
		console.log(newState.form[name].valid)
		this.setState(newState);
		return newState.form[name].valid
	}

	handleUserInput(event) {
		const name = event.target.name;
		const value = event.target.value;
		let newState = this.state
		newState.form[name].value = value
		this.setState(newState, () => { this.validateField(name, value) });
	}

	componentDidMount() {
		this.imagesSaved = this.imagesSaved.bind(this);
		this.handleSubmit = this.handleSubmit.bind(this);
		this.imagesSaved();
	}

	render() {
		return (
			<div className="App">
				<form action="#" onSubmit={this.handleSubmit} noValidate>
					<label htmlFor="backgroundsCount">
						Number of backgrounds
					</label>
					<br/>
					<input type="number" name="backgroundsCount" value={this.state.form.backgroundsCount.value} onChange={this.handleUserInput} required/>
					<p>{this.state.form.backgroundsCount.errMsg}</p>
					<br/>
					<label htmlFor="width">
						Width
					</label>
					<br/>
					<input type="number" name="imagesWidth" value={this.state.form.imagesWidth.value} onChange={this.handleUserInput} required/>
					<p>{this.state.form.imagesWidth.errMsg}</p>
					<br/>
					<label htmlFor="height">
						Height
					</label>
					<br/>
					<input type="number" name="imagesHeight" value={this.state.form.imagesHeight.value} onChange={this.handleUserInput} required/>
					<p>{this.state.form.imagesHeight.errMsg}</p>
					<br/>
					<label>
						Download Path
					</label>
					<br/>
					<input type="text" name="downloadPath" value={this.state.form.downloadPath.value} onChange={this.handleUserInput} required/>
					<br/>
					<input type="submit" value="Get New Images"/>
				</form>
				<p>
					{this.state.imagesSaved.toString()}
				</p>
				<p>
					{this.state.responseMsg.toString()}
				</p>
			</div>
		);
	}
}

export default ImagesRetriever;
