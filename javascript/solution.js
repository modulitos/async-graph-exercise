const fetch = require("node-fetch");
const assert = require("assert");

const HOST = "localhost";
const PORT = 7878;

const STARTING_NODE = 1;

const main = async () => {
  let nodeIdsToGet = [STARTING_NODE];
  let nodeIdsRetrieved = new Set();
  let promises = [];
  let total = 0;

  while (true) {
    if (nodeIdsToGet.length > 0) {
      // create a new promise, add to end of promises queue
      const numNodesToProcess = nodeIdsToGet.length;
      nodesToProcess = nodeIdsToGet.slice(0, numNodesToProcess);
      nodeIdsToGet = nodeIdsToGet.slice(numNodesToProcess);

      const nextPromises = nodesToProcess.map(async nextNodeId => {
        // TODO: check if nextNodeIds is already in nodeIdsRetrieved.
        const res = await fetch(`http://${HOST}:${PORT}/node/${nextNodeId}`);
        const body = await res.json();
        total += body.reward;
        return body.children;
      });
      promises = promises.concat(nextPromises);
    }
    if (promises.length > 0) {
      // pop off promise from front of queue, await it.
      let nextPromise = promises.shift();
      // TODO: we can afford to wait on this node, when there are
      // other requests to process... Is there a way to fetch all of
      // our promises, and process them as they return? (not wait for
      // all of them to return)
      let nextNodeIds = await nextPromise;
      nextNodeIds.forEach(child => nodeIdsToGet.push(child));
    }
    if (promises.length === 0 && nodeIdsToGet.length == 0) {
      return total;
    }
  }
};

console.log("starting...");
main()
  .then(total => {
    console.log("result:", total);
    // TODO: this should take 4 cycles, each taking [1, 1, 5, 5] seconds to process
    assert.equal(total, 3850);
  })
  .catch(err => {
    console.log("program rejected with err:", err);
  });
