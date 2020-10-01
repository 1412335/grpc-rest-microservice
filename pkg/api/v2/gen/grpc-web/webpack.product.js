const path = require('path')

module.exports = {
    mode: "production",
    entry: "./cmd/client.js",
    output: {
        filename: 'main.js',
        path: path.resolve(__dirname, 'dist')
    },
    // watch: true,
};