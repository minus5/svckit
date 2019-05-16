const path = require('path');

module.exports = {
  entry: './src/main.js',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'sdk.js',
    library: ['minus5', 'api'],
    libraryTarget: 'window'
  },
  mode: "development"
};
