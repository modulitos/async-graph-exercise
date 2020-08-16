const fetch = require("node-fetch");
const assert = require("assert");

const HOST = "localhost";
const PORT = 7878;

const STARTING_NODE = 1;

const main = async () => {
  let nodeIdsToGet = [STARTING_NODE];
  let total = 0;

  while (nodeIdsToGet.length > 0) {
    console.log("executing cycle, running total:", total);

    const nextResponsePromises = nodeIdsToGet.map(async id => {
      const res = await fetch(`http://${HOST}:${PORT}/node/${id}`);
      return res;
    });
    const responses = await Promise.all(nextResponsePromises);
    const responseBodyPromises = responses.map(async response => {
      // TODO: filter our responses with non-200 return codes.
      return await response.json();
    });
    const responseBodies = await Promise.all(responseBodyPromises);
    nodeIdsToGet = responseBodies.reduce((acc, body) => {
      total += body.reward;
      return acc.concat([body.left, body.right].filter(el => el !== null));
    }, []);
  }
  return total;
};

console.log("starting...");
main()
  .then(total => {
    console.log("result:", total);
    // this should take 4 cycles, each taking [1, 1, 5, 5] seconds to process

    // TODO: should take less than 10 seconds to process
    assert.equal(total, 3850);
  })
  .catch(err => {
    console.log("program rejected with err:", err);
  });
