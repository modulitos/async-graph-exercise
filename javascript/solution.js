const fetch = require("node-fetch");
const assert = require("assert");

const HOST = "localhost";
const PORT = 7878;

const STARTING_NODE = 1;

const main = async () => {
  let nodeIdsToGet = [STARTING_NODE];
  let nodeIdsRetrieved = new Set();
  let idToPromises = {};
  let total = 0;

  while (true) {
    if (nodeIdsToGet.length > 0) {
      const nextPromises = nodeIdsToGet
        .map(async nodeId => {
          // TODO: check if nextNodeIds is already in nodeIdsRetrieved.
          const res = await fetch(`http://${HOST}:${PORT}/node/${nodeId}`);
          const body = await res.json();
          total += body.reward;
          // return body.children;
          body.children.forEach(child => nodeIdsToGet.push(child));
          nodeIdsRetrieved.add(nodeId);
          delete idToPromises[nodeId];
          return nodeId;
        })
        .reduce((acc, promise, i) => {
          acc[nodeIdsToGet[i]] = promise;
          return acc;
        }, {});
      nodeIdsToGet = [];
      Object.assign(idToPromises, nextPromises);
    } else if (Object.values(idToPromises).length > 0) {
      const nodeId = await Promise.race(Object.values(idToPromises));
      console.log("promise.race result, nodeId:", nodeId);
    } else {
      await Promise.all(Object.values(idToPromises));
      return total;
    }
  }
};

console.log("starting...");
main()
  .then(total => {
    console.log("result:", total);
    // To process most efficiently, this will have the following requests in flight:

    // | time (sec) | nodes being processed           |
    // |------------+---------------------------------|
    // |          1 | 1                               |
    // |          2 | 5, 3                            |
    // |          3 | 6, 7, 9, 10                     |
    // |          4 | 6, 101, 102, 103, 104, 105, 106 |
    // |          5 | 6, 101, 102, 103, 104, 105, 106 |
    // |          6 | 6, 101, 102, 103, 104, 105, 106 |
    // |          7 | 6, 101, 102, 103, 104, 105, 106 |
    // |          8 | 101, 102, 103, 104, 105, 106    |

    assert.equal(total, 3850);
  })
  .catch(err => {
    console.log("program rejected with err:", err);
  });
