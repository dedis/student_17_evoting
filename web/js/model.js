// jshint esversion: 6

const messages = {
    'nested': {
	'ServerIdentity': {
	    'fields': {
		'Public': {
		    'rule': 'optional',
		    'type': 'bytes',
		    'id': 1
		},
		'Id': {
		    'rule': 'optional',
		    'type': 'bytes',
		    'id': 2
		},
		'Address': {
		    'rule': 'required',
		    'type': 'string',
		    'id': 3
		},
		'Description': {
		    'rule': 'optional',
		    'type': 'string',
		    'id': 4
		}
	    }
	},
	'Roster': {
	    'fields': {
		'Id': {
		    'rule': 'optional',
		    'type': 'bytes',
		    'id': 1
		},
		'List': {
		    'rule': 'repeated',
		    'type': 'ServerIdentity',
		    'id': 2
		},
		'Aggregate': {
		    'rule': 'optional',
		    'type': 'bytes',
		    'id': 3
		}
	    }
	},
	'GenerateRequest': {
	    'fields': {
		'Name': {
		    'rule': 'required',
		    'type': 'string',
		    'id': 1
		},
		'Roster': {
		    'rule': 'required',
		    'type': 'Roster',
		    'id': 2
		}
	    }
	},
	'GenerateResponse': {
	    'fields': {
		'Key': {
		    'rule': 'required',
		    'type': 'bytes',
		    'id': 1
		},
		'Hash': {
		    'rule': 'required',
		    'type': 'bytes',
		    'id': 2
		}
	    }
	},
	'CountResponse': {
	    'fields': {
		'Count': {
		    'rule': 'required',
		    'type': 'sint32',
		    'id': 1
		}
	    }
	},
	'ClockRequest': {
	    'fields': {
		'Roster': {
		    'rule': 'required',
		    'type': 'Roster',
		    'id': 1
		}
	    }
	},
	'ClockResponse': {
	    'fields': {
		'Time': {
		    'rule': 'required',
		    'type': 'double',
		    'id': 1
		},
		'Children': {
		    'rule': 'required',
		    'type': 'sint32',
		    'id': 2
		}
	    }
	},
	'CountRequest': {
	    'fields': {
		
	    }
	}
    }
};
