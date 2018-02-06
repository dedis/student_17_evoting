package service

// func TestAssertLevel(t *testing.T) {
// 	log := map[string]*stamp{"0": &stamp{0, true, 0}, "1": &stamp{1, false, 0}}
// 	service := &Service{state: &state{log}}

// 	// Not logged in
// 	_, err := service.assertLevel("2", false)
// 	assert.NotNil(t, err)

// 	// Not priviledged
// 	_, err = service.assertLevel("1", true)
// 	assert.NotNil(t, err)

// 	// Valid assertion
// 	user, _ := service.assertLevel("0", true)
// 	assert.Equal(t, 0, int(user))
// 	user, _ = service.assertLevel("1", false)
// 	assert.Equal(t, 1, int(user))
// }
