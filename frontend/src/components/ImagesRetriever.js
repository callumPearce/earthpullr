import React from 'react';
import * as Wails from '@wailsapp/runtime'
import PropTypes from 'prop-types';
import { FormControl, Container, Box, Grid, TextField, Tooltip, Button, Typography, CircularProgress } from '@mui/material';
import { styled } from '@mui/material/styles';

const MAX_RES = 7680
const MAX_BACKGROUND_COUNT = 50

const imageDimensionValidation = dim => {
	if (!dim) {
		return "Both background width and height must be specified"
	}
	let parsedDim = parseFloat(dim);
	if (!((parsedDim | 0) === parsedDim)) {
		return "Dimension must be a number"
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

const CssTextField = styled(TextField)({
	'& label.Mui-focused': {
		color: 'white',
	},
	'& label': {
		color: '#1976d2',
	},
	'&:hover label': {
		color: 'white',
	},
	'& .MuiInput-underline:after': {
		borderBottomColor: 'red',
	},
	'& .MuiOutlinedInput-root': {
		'& fieldset': {
			borderColor: '#125394',
		},
		'&.Mui-focused input': {
			color: 'white',
		},
		'& input': {
			color: '#1976d2',
		},
		'&:hover fieldset': {
			borderColor: 'white',
		},
		'&.Mui-focused fieldset': {
			borderColor: 'white',
		},
	},
});

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
					<Typography variant="caption" component="div" color="#1976d2">
						{props.label}
					</Typography>
				</Box>
			</Box>
		);
	}

	StyledAdornment(props) {
		return (
			<p style={{color: "#125394"}}>
				px
			</p>
		)
	}

	render() {
		return (
			<Container component="div" className="App" sx={{  alignItems: 'center', justifyContent: 'center'}}>
				{this.state.displayForm && <FormControl action="#" noValidate>
					<Grid container spacing={2}>
						<Grid item xs={12}>
							<Typography variant="overline" style={{color: "#1976d2", fontSize: "1.5ch"}}>earthpull/r</Typography>
						</Grid>
						<Grid item xs={12}>
							<CssTextField
								type="text"
								label="Download Path"
								name="downloadPath"
								value={this.state.form.downloadPath.value}
								onChange={this.handleUserInput}
								error={!this.state.form.downloadPath.valid}
								helperText={this.state.form.downloadPath.errMsg}
								sx={{width: "24.5ch", m: .5}}
								required
							/>
						</Grid>
						<Grid item xs={12}>
							<Tooltip title="Initially set to your current screen's width">
								<CssTextField
									type="tel"
									label="Width"
									name="imagesWidth"
									value={this.state.form.imagesWidth.value}
									onChange={this.handleUserInput}
									error={!this.state.form.imagesWidth.valid}
									helperText={this.state.form.imagesWidth.errMsg}
									InputProps={{endAdornment: <this.StyledAdornment/>}}
									sx={{width: "12ch", m: .5}}
									required
								/>
							</Tooltip>
							<Tooltip title="Initially set to your current screen's height">
								<CssTextField
									type="tel"
									label="Height"
									name="imagesHeight"
									value={this.state.form.imagesHeight.value}
									onChange={this.handleUserInput}
									error={!this.state.form.imagesHeight.valid}
									helperText={this.state.form.imagesHeight.errMsg}
									InputProps={{endAdornment: <this.StyledAdornment/>}}
									sx={{width: "12ch", m: .5}}
									required
								/>
							</Tooltip>
						</Grid>
						<Grid item xs={12}>
							<CssTextField
								type="tel"
								label="Images"
								name="backgroundsCount"
								value={this.state.form.backgroundsCount.value}
								onChange={this.handleUserInput}
								error={!this.state.form.backgroundsCount.valid}
								helperText={this.state.form.backgroundsCount.errMsg}
								inputProps={{min: 0, style: {textAlign: 'center'}}}
								sx={{width: "24.5ch", m: .5}}
								required
							/>
						</Grid>
						<Grid item xs={12}>
							<Button
								type="submit"
								variant="contained"
								onClick={this.handleSubmit}
								back
							>
								Pull
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
					<Box sx={{ padding: "2ch", position: 'relative'}}>
						<Button
							variant="contained"
							onClick={() => this.setDisplayFormState("")}
						>
							Done
						</Button>
					</Box>
				}
			</Container>
		);
	}
}

export default ImagesRetriever;
