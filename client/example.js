const routes = require('./routes/client');

const client = routes.Client('http://localhost:8000');

// Display loadbalancers
client.list()
    .then(lists => {
        console.log('\nDisplay loadbalancers\n')
        for (const obj of lists) {
            console.log(obj)
        }
        console.table(lists)
    })
    .catch(err => {
        console.error(err)
    })
    .finally(()=> {
        console.log('Finished')
    })

// Update machine
client.put({isWork: 1, id: 2})
    .then(res => {
        console.log('Update')
        console.log(`Response: ${res}`)
        return client.list()
                .then(lists => {
                    console.table(lists)
                })
    })
    .catch(err => {
        console.error(err)
    })  