// jshint esversion: 6 

$(() => {
    'use strict';

    const proto = protobuf.Root.fromJSON(messages);
    
    let roster = null;
    let current = null;
    let elections = {};

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
	$('#modal-key').text(elections[current].key);
    });

    $('#modal-add').click(() => {
	elections[current].cast(new UUID(1).toString()).then((data) => {
	    populate('#ballots', elections[current].ballots);
	    $.notify('Vote casted', 'success');
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

	const election = new Election(name, roster, proto);
	election.generate().then(() => {
	    const number = Object.keys(elections).length;
	    const row = (number / 3) | 0;
	    const col = number % 3;

            if (col == 0)
	    	$('.grid').append('<tr><td></td><td></td><td></td></tr>');
        
            $(`.grid tr:eq(${row}) td:eq(${col})`).append(`<div class="cell">${name}</div>`);
            $(`.grid tr:eq(${row}) td:eq(${col}) > .cell`).css('background-color', color());
            $(`.grid tr:eq(${row}) td:eq(${col}) > .cell`).attr('data-toggle', 'modal');
            $(`.grid tr:eq(${row}) td:eq(${col}) > .cell`).attr('data-target', '#modal');

	    elections[name] = election;
	    $.notify(`Election with key ${election.key} generated`, 'success');
	}).catch((error) => {
	    console.log(error.toString());
	    $.notify(error, 'error');
	});
    });
});
