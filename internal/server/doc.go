package server

// the server has 3 state, and each state decides different responsibitliies
// of the server

// 1. the head node reponsible for:
// * set of key/value
// * get of key/value
// * assign or increment version

// 2. the middle node
// * only get of key/value
// * forward the request to tail if current version of key is dirty
// * send the ack back to the precessor

// 3. the tail node reponsible for:
//
// * version query
