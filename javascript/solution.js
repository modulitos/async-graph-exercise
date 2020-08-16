const fetch = require("node-fetch");
const assert = require("assert");

const HOST = "localhost";
const PORT = 7878;

const main = async () => {
  console.log("fetching...");
  const res = await fetch(`http://${HOST}:${PORT}/node/1`);
  const body = await res.json();
  assert.equal(body.score, 100);
  return body.score;
};

console.log("starting...");
main()
  .then(totalScore => {
    console.log("result:", totalScore);
  })
  .catch(err => {
    console.log("program rejected with err:", err);
  });
