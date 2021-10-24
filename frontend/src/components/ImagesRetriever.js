import React from 'react';
import * as Wails from '@wailsapp/runtime'
import PropTypes from 'prop-types';
import { FormControl, Container, Box,	 Grid, TextField, InputAdornment, Button, Typography, CircularProgress } from '@mui/material';

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
				backgroundsCount: {value: null, errMsg: "", valid: true, validator: backgroundCountValidation},
				imagesWidth: {value: window.screen.width * window.devicePixelRatio, errMsg: "", valid: true, validator: imageDimensionValidation},
				imagesHeight: {value: window.screen.height * window.devicePixelRatio, errMsg: "", valid: true, validator: imageDimensionValidation},
				downloadPath: {value: "", errMsg: "", valid: true, validator: downloadPathValidation},
			},
			imagesSaved: 0,
			responseMsg: "",
			displayForm: true,
			displayProgressBar: false,
			displayGetMoreImages: false
		};

		this.ProgressBar.propTypes = {
			/**
			 * The value of the progress indicator for the determinate and buffer variants.
			 * Value between 0 and 100.
			 */
			value: PropTypes.number.isRequired,
		};

		this.handleSubmit = this.handleSubmit.bind(this);
		this.setDisplayFormState = this.setDisplayFormState.bind(this);
		this.handleUserInput = this.handleUserInput.bind(this);
	}

	setDisplayFormState(errMsg) {
		if (errMsg != null) {
			this.setState({
				displayForm: true,
				displayProgressBar: false,
				displayGetMoreImages: false,
				responseMsg: errMsg,
				imagesSaved: 0
			})
		} else {
			this.setState({
				displayForm: true,
				displayProgressBar: false,
				displayGetMoreImages: false,
				responseMsg: "",
				imagesSaved: 0
			})
		}
	}

	setDisplayProgressBarState() {
		this.setState({
			displayForm: false,
			displayProgressBar: true,
			displayGetMoreImages: false
		})
	}

	setDisplayGetMoreImagesState() {
		this.setState({
			displayForm: false,
			displayProgressBar: true,
			displayGetMoreImages: true
		})
	}

	imagesSaved () {
		Wails.Events.On("image_saved", imagesSaved => {
			this.setState({ imagesSaved: this.state.imagesSaved + 1 });
		});
	}

	handleSubmit() {
		let valid = true;
		for (const [name, value] of Object.entries(this.state.form)) {
			valid = this.validateField(name, value.value) && valid
		}
		if (valid) {
			this.setDisplayProgressBarState();
			const request = {
				BackgroundsCount: parseInt(this.state.form.backgroundsCount.value),
				Width: parseInt(this.state.form.imagesWidth.value),
				Height: parseInt(this.state.form.imagesHeight.value),
				DownloadPath: this.state.form.downloadPath.value
			}
			window.backend.BackgroundRetriever.GetBackgrounds(request).then(result =>
				this.setDisplayGetMoreImagesState()
			)
			.catch(err =>
				this.setDisplayFormState(err)
			)
		}
		else {
			this.setState({responseMsg: "Form is invalid"})
		}
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

	get progress_bar_value() {
		if (this.state.form.backgroundsCount.value != null)
			return (this.state.imagesSaved/this.state.form.backgroundsCount.value)*100
		else
			return 0
	}

	get progress_bar_label() {
		if (this.state.form.backgroundsCount.value !=  null)
			return `${this.state.imagesSaved}/${this.state.form.backgroundsCount.value}`
		else
			return ''
	}

	ProgressBar(props) {
		return (
			<Box sx={{ position: 'relative' }}>
				<CircularProgress size="10ch" variant="determinate" {...props} />
				<Box
					sx={{
						top: 0,
						left: 0,
						bottom: 0,
						right: 0,
						position: 'absolute',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
					}}
				>
					<Typography variant="caption" component="div" color="text.secondary">
						{props.label}
					</Typography>
				</Box>
			</Box>
		);
	}

	render() {
		return (
			<Container component="div" className="App" sx={{ width: '25ch' }}>
				{this.state.displayForm && <FormControl action="#" noValidate>
					<Grid container spacing={2}>
						<Grid item xs={12}>
							<TextField
								type="text"
								label="Download Path"
								name="downloadPath"
								value={this.state.form.downloadPath.value}
								onChange={this.handleUserInput}
								error={!this.state.form.downloadPath.valid}
								helperText={this.state.form.downloadPath.errMsg}
								sx={{width: "14.5ch", m: .5}}
								required
							/>
						</Grid>
						<Grid item xs={12}>
							<TextField
								type="number"
								label="Width"
								name="imagesWidth"
								value={this.state.form.imagesWidth.value}
								onChange={this.handleUserInput}
								error={!this.state.form.imagesWidth.valid}
								helperText={this.state.form.imagesWidth.errMsg}
								InputProps={{endAdornment: <InputAdornment position="end">px</InputAdornment>}}
								sx={{width: "7ch", m: .5}}
								required
							/>
							<TextField
								type="number"
								label="Height"
								name="imagesHeight"
								value={this.state.form.imagesHeight.value}
								onChange={this.handleUserInput}
								error={!this.state.form.imagesHeight.valid}
								helperText={this.state.form.imagesHeight.errMsg}
								InputProps={{endAdornment: <InputAdornment position="end">px</InputAdornment>}}
								sx={{width: "7ch", m: .5}}
								required
							/>
						</Grid>
						<Grid item xs={12}>
							<TextField
								type="number"
								label="Backgrounds"
								name="backgroundsCount"
								value={this.state.form.backgroundsCount.value}
								onChange={this.handleUserInput}
								error={!this.state.form.backgroundsCount.valid}
								helperText={this.state.form.backgroundsCount.errMsg}
								inputProps={{min: 0, style: {textAlign: 'center'}}}
								sx={{width: "14.5ch", m: .5}}
								required
							/>
						</Grid>
						<Grid item xs={12}>
							<Button
								type="submit"
								variant="contained"
								onClick={this.handleSubmit}
							>
								Get New Images
							</Button>
						</Grid>
						<Grid item xs={12}>
							<Typography variant="body1" component="div" color="red">
								{this.state.responseMsg.toString()}
							</Typography>
						</Grid>
					</Grid>
				</FormControl>
				}
				{ this.state.displayProgressBar &&
				<this.ProgressBar
					hidden={this.state.displayProgressBar}
					value={this.progress_bar_value}
					label={this.progress_bar_label}/>
				}
				{ this.state.displayGetMoreImages &&
					<Box sx={{ position: 'relative'}}>
						<Button
							variant="contained"
							onClick={() => this.setDisplayFormState("")}
						>
							Get More Images
						</Button>
					</Box>
				}
			</Container>
		);
	}
}

export default ImagesRetriever;
