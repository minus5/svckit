const path = require('path');


module.exports = {
  entry: './src/main.js',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'sdk.js',
    library: 'mnu5',
    libraryTarget: 'window'
  },
  mode: "development"
};
