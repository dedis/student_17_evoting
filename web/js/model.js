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
	'Point': {
	    'fields': {
		'X': {
		    'rule': 'required',
		    'type': 'bytes',
		    'id': 1
		},
		'Y': {
		    'rule': 'required',
		    'type': 'bytes',
		    'id': 2
		},
		'Z': {
		    'rule': 'optional',
		    'type': 'bytes',
		    'id': 3
		}
	    }
	},
	'Ballot': {
	    'fields': {
		'Alpha': {
		    'rule': 'required',
		    'type': 'bytes',
		    'id': 1
		},
		'Beta': {
		    'rule': 'required',
		    'type': 'bytes',
		    'id': 2
		}
	    }
	},
	'CastRequest': {
	    'fields': {
		'Election': {
		    'rule': 'required',
		    'type': 'string',
		    'id': 1
		},
		'Ballot': {
		    'rule': 'required',
		    'type': 'Ballot',
		    'id': 2
		}
	    }
	},
	'CastResponse': {
	    'fields': {
	    }
	},
	'ShuffleRequest': {
	    'fields': {
		'Election': {
		    'rule': 'required',
		    'type': 'string',
		    'id': 1
		}
	    }
	},
	'ShuffleResponse': {
	    'fields': {
	    }
	},
	'FetchRequest': {
	    'fields': {
		'Election': {
		    'rule': 'required',
		    'type': 'string',
		    'id': 1
		},
		'Block': {
		    'rule': 'required',
		    'type': 'int32',
		    'id': 2
		}
	    }
	},
	'FetchResponse': {
	    'fields': {
		'Ballots': {
		    'rule': 'repeated',
		    'type': 'Ballot',
		    'id': 1
		}
	    }
	},
	'DecryptionRequest': {
	    'fields': {
		'Election': {
		    'rule': 'required',
		    'type': 'string',
		    'id': 1
		}
	    }
	},
	'DecryptionResponse': {
	    'fields': {
	    }
	}
    }
};
