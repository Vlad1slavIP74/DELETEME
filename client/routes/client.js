const http = require('../common/http');

const Client = baseUrl => {
    const client = http.Client(baseUrl);

    return {
        list: () => client.get('/list'),
        put: (data) => client.put('/update', data)
    }
};

module.exports = { Client };