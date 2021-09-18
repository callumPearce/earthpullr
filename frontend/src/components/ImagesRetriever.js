import React, { useState } from 'react';
import * as Wails from '@wailsapp/runtime'

class ImagesRetriever extends React.Component {
	constructor(props, context) {
		super(props);

		this.state = {
			images_saved: 0,
			images_required : 0
		};
	}

	imagesSaved () {
		Wails.Events.On("image_saved", images_saved => {
			this.setState({ images_saved: this.state.images_saved + 1 });
		});
	}

	handleGetNewImages() {
		window.backend.BackgroundRetriever.GetBackgrounds().then(result =>
			console.log(result)
		);
	}

	componentDidMount() {
		this.imagesSaved = this.imagesSaved.bind(this);
		this.handleGetNewImages = this.handleGetNewImages.bind(this)
		this.imagesSaved();
	}

	render() {
		return (
			<div className="App">
				<form>
					<label htmlFor="backgrounds_count">
						Backgrounds Count
					</label>
					<br/>
					<input type="number" id="backgrounds_count" name="backgrounds_count">
					</input>
					<br/>
					<label htmlFor="width">
						Width
					</label>
					<br/>
					<input type="number" id="width" name="width">
					</input>
					<br/>
					<label htmlFor="height">
						Height
					</label>
					<br/>
					<input type="number" id="height" name="height">
					</input>
				</form>
				<button onClick={() => this.handleGetNewImages()} type="button">
					Get New Images
				</button>
				<p>
					{this.state.images_saved.toString()}
				</p>
			</div>
		);
	}
}

export default ImagesRetriever;
