const path = require('path')

module.exports = {
    mode: "development",
    entry: "./cmd/client.js",
    output: {
        filename: 'main.js',
        path: path.resolve(__dirname, 'dist')
    },
    devServer: {
        port: 8081, // use any port suitable for your configuration
        host: '0.0.0.0', // to accept connections from outside container
        publicPath: '/dist/',
        watchOptions: {
            aggregateTimeout: 500, // delay before reloading
            poll: 1000 // enable polling since fsevents are not supported in docker
        }
    }
};