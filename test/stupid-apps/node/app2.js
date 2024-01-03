// app.js
const { fork } = require('child_process');

const subprocess1Handle = fork('subprocess1.js');
const subprocess2Handle = fork('subprocess2.js');

subprocess1Handle.on('exit', () => {
    subprocess2Handle.kill();
});

subprocess2Handle.on('exit', () => {
    console.log("Bye!");
});
