const path = require('path');
const common = require('./webpack.common');
const merge = require('webpack-merge');
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = merge(common, {
    mode: "development",
    // devtool: "none",
    output: {
        filename: "[name].bundle.js",
        path: path.resolve(__dirname, "dist")
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: "./src/template.html",
        })
    ],
    module: {
        rules: [
            {
                test: /\.scss$/,
                use: [
                    "style-loader", //3. Inject styles into DOM
                    "css-loader", //2. Turns css into commonjs
                    "sass-loader" //1. Turns sass into css
                ]
            }
        ]
    },
    devServer: {
        port: 8081, // use any port suitable for your configuration
        host: '0.0.0.0', // to accept connections from outside container
        // publicPath: '/dist/',
        watchOptions: {
            aggregateTimeout: 500, // delay before reloading
            poll: 1000 // enable polling since fsevents are not supported in docker
        }
    }
});