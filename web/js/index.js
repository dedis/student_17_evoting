// jshint esversion: 6 

$(() => {
    'use strict';

    const proto = protobuf.Root.fromJSON(messages);
    const curve = elliptic.ec;
    
    let roster = null;
    let current = null;
    let elections = {};
    let counter = 0;

    $.notify.defaults({position: "top left", autoHideDelay: 2000});

    $('#file-button').on('click', () => {
	$('#file-input').trigger('click');
    });

    $('#file-input').change(() => {
	readFile($('#file-input')[0])
	    .then(parseRoster)
	    .then(result => {
	        $('#file-form').attr('placeholder', $('#file-input').val().split('\\').pop());
		roster = result;
		$.notify(`Roster uploaded`, 'success');
	    }).catch((error) => {
		$.notify(error.message, 'error');
	    });
    });

    $('.grid').on('click', 'tr', 'td > .cell', (event) => {
	current = $(event.target).text();
	populate('#ballots', elections[current].ballots);
	$('#modal-election').text(`Election - ${current}`);
	$('#modal-genesis').text(elections[current].hash);
	$('#modal-key').text(elections[current].key.x.toString(16, 2));
	$('#modal-roster').empty();
	$.each(elections[current].roster.servers, (index, element) => {
	    let button = `<button id="modal-node" type="button" class="btn btn-secondary">
                            ${element.Address}
                          </button>`;
	    $('#modal-roster').append(button); 
	});
    });

    $('#modal-ballots').click(() => {
	populate('#ballots', elections[current].ballots);
    });

    $('#modal-roster').on('click', '#modal-node', (event) => {
	if (!elections[current].shuffled) {
	    $.notify('Election not yet shuffled', 'warn');
	    return;
	}
	
	let node = $(event.target).text().trim();
	elections[current].fetch(node).then(() => {
	    populate('#ballots', elections[current].shuffles);
	});
    });

    $('#modal-add').click(() => {
	if (elections[current].shuffled) {
	    $.notify('Cannot add ballot after shuffling', 'warn');
	    return;
	}

	elections[current].cast('ballot'+counter++).then((data) => {
	    populate('#ballots', elections[current].ballots);
	    $.notify('Vote casted', 'success');
	}).catch((error) => {
	    $.notify(error.message, 'error');
	});
    });

    $('#modal-shuffle').click(() => {
	if (elections[current].shuffled) {
	    $.notify('Ballots already shuffled', 'warn');
	    return;
	}

	elections[current].shuffle().then(() => {
	    $.notify('Successfully shuffled ballots', 'success');
	}).catch((error) => {
	    $.notify(error.message, 'error');
	});
    });

    $('#modal-decrypt').click(() => {
	elections[current].decrypt().then(() => {
	    $.notify('Successful decryption', 'success');
	}).catch((error) => {
	    $.notify(error.message, 'error'); 
	});
    });

    $('#election-input').keypress((event) => {
	if (event.keyCode != 13)
	    return;

	if (roster == null) {
	    $.notify('Please upload roster TOML', 'error');
	    return;
	}

	const name = $('#election-input').val();
	if (name.length == 0) {
	    $.notify('Please specify election name', 'error');
	    return;
	} else if (elections[name] != undefined) {
	    $.notify(`Election ${name} already exists`, 'error');
	    return;
	}

	const election = new Election(name, roster, proto, new curve('ed25519'));
	election.generate().then(() => {
            const number = $('.grid tr td > .cell').length;
            const row = (number / 3) | 0;
            const col = number % 3;
            
            if (col == 0)
            	$('.grid').append('<tr><td></td><td></td><td></td></tr>');
            
            $(`.grid tr:eq(${row}) td:eq(${col})`).append(`<div class="cell">${name}</div>`);
            $(`.grid tr:eq(${row}) td:eq(${col}) > .cell`).css('background-color', color());
            $(`.grid tr:eq(${row}) td:eq(${col}) > .cell`).attr('data-toggle', 'modal');
            $(`.grid tr:eq(${row}) td:eq(${col}) > .cell`).attr('data-target', '#modal');
	    
	    elections[name] = election;
	    $.notify(`Election generated`, 'success');
	}).catch((error) => {
	    $.notify(error.message, 'error');
	});
    });
});
