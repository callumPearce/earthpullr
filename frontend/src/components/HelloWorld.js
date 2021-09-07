import React, { useState } from 'react';
import Modal from 'react-modal';

function HelloWorld() {
	const [showModal, setShowModal] = useState(false);
	const [result, setResult] = useState(null);

	function handleGetNewImages() {
		window.backend.RetrieveImages().then(result =>
			console.log(result)
		);
	}
	
	return (
		<div className="App">
			<button onClick={() => handleGetNewImages()} type="button">
				Get New Images
      		</button>
		</div>
	);
}

export default HelloWorld;
