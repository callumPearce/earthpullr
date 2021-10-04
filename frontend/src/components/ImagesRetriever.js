import React, { useState } from 'react';
import * as Wails from '@wailsapp/runtime'

class ImagesRetriever extends React.Component {
	constructor(props, context) {
		super(props);

		this.state = {
			imagesSaved: 0,
			imagesRequired : 0,
			imagesWidth: 0,
			imagesHeight: 0,
			responseMsg: ""
		};

		this.handleSubmit = this.handleSubmit.bind(this);
		this.handleImagesRequiredChange = this.handleImagesRequiredChange.bind(this);
		this.handleImagesWidthChange = this.handleImagesWidthChange.bind(this);
		this.handleImagesHeightChange = this.handleImagesHeightChange.bind(this);
	}

	imagesSaved () {
		Wails.Events.On("image_saved", imagesSaved => {
			this.setState({ imagesSaved: this.state.imagesSaved + 1 });
		});
	}

	handleSubmit(event) {
		var request = {
			ImagesRequired: parseInt(this.state.imagesRequired),
			Width: parseInt(this.state.imagesWidth),
			Height: parseInt(this.state.imagesHeight)
		}
		window.backend.BackgroundRetriever.GetBackgrounds(request).
		then(result =>
			this.setState({ responseMsg: result})
		).catch(err =>
			this.setState({ responseMsg: err})
		)
		event.preventDefault()
	}

	handleImagesRequiredChange(event) {
		this.setState({imagesRequired : event.target.value});
	}

	handleImagesWidthChange(event) {
		this.setState({imagesWidth : event.target.value});
	}

	handleImagesHeightChange(event) {
		this.setState({imagesHeight: event.target.value});
	}

	componentDidMount() {
		this.imagesSaved = this.imagesSaved.bind(this);
		this.handleSubmit = this.handleSubmit.bind(this);
		this.imagesSaved();
	}

	render() {
		return (
			<div className="App">
				<form action="#" onSubmit={this.handleSubmit}>
					<label htmlFor="backgroundsCount">
						Backgrounds Count
					</label>
					<br/>
					<input type="number" name="images_required" value={this.state.imagesRequired} onChange={this.handleImagesRequiredChange}>
					</input>
					<br/>
					<label htmlFor="width">
						Width
					</label>
					<br/>
					<input type="number" name="width" value={this.state.imagesWidth} onChange={this.handleImagesWidthChange}>
					</input>
					<br/>
					<label htmlFor="height">
						Height
					</label>
					<br/>
					<input type="number" name="height" value={this.state.imagesHeight} onChange={this.handleImagesHeightChange}>
					</input>
					<br/>
					<input type="submit" value="Get New Images">
					</input>
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
