// jshint esversion: 6

function readFile(input) {
    if (!input.files || !input.files[0])
	throw 'File not found';

    const file = input.files[0];
    const reader = new FileReader();

    return new Promise((resolve, reject) => {
	reader.onload = () => {
	    resolve(reader.result);
	};

	reader.onerror = (error) => {
	    reject(error);
	};

	reader.readAsText(file);
    });
}

function color() {
    const letters = '0123456789ABCDEF';
    let color = '#';
    for (let i = 0; i < 6; i++)
        color += letters[Math.floor(Math.random() * 16)];

    return color;
}

function populate(table, data) {
    $(`${table} tbody tr`).remove();
    $.each(data, (index, element) => {
	let alpha = misc.uint8ArrayToHex(element.Alpha);
	let beta = misc.uint8ArrayToHex(element.Beta);
	$(table).append(`<tr><td>${alpha}<br>${beta}</td></tr>`);
    });
}
