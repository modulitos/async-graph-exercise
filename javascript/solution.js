const fetch = require("node-fetch");
const assert = require("assert");

const HOST = "localhost";
const PORT = 7878;

const STARTING_NODE = 1;

// This solution leverages concurrency to calculate the solution in
// the shortest amount of time possible. It calculates the result in 8
// seconds, assuming no additional network latency. To process our
// nodes most efficiently, we will have the following requests in
// flight:

// | time (sec) | nodes requests being processed  |
// |------------+---------------------------------|
// |          1 | 1                               |
// |          2 | 5, 3                            |
// |          3 | 6, 7, 9, 10                     |
// |          4 | 6, 101, 102, 103, 104, 105, 106 |
// |          5 | 6, 101, 102, 103, 104, 105, 106 |
// |          6 | 6, 101, 102, 103, 104, 105, 106 |
// |          7 | 6, 101, 102, 103, 104, 105, 106 |
// |          8 | 101, 102, 103, 104, 105, 106    |

const main = async () => {
  // let nodeIdsToGet = [STARTING_NODE];
  let nodeIdsToGet = new Set([STARTING_NODE]);
  let nodeIdsRetrieved = new Set();
  let idToPromises = {};
  let total = 0;

  while (true) {
    if (nodeIdsToGet.size > 0) {
      const idToNextPromises = [...nodeIdsToGet].reduce((acc, nodeId) => {
        // delete the item from nodeIdsToGet, as it has been converted to a promise at this point.
        nodeIdsToGet.delete(nodeId);
        const promise = (async nodeId => {
          if (nodeIdsRetrieved.has(nodeId)) {
            // Skip processing nodes that have already been fetched.
            return null;
          }
          const res = await fetch(`http://${HOST}:${PORT}/node/${nodeId}`);
          const body = await res.json();
          // Increment the reward:
          total += body.reward;
          // Add the node's children to the queue:
          body.children.forEach(child => nodeIdsToGet.add(child));
          // Mark the node as having already been retrieved.
          nodeIdsRetrieved.add(nodeId);
          // Remove the fetched node from our in-flight promises:
          delete idToPromises[nodeId];
          return nodeId;
        })(nodeId);
        acc[nodeId] = promise;
        return acc;
      }, {});

      // All of the pending node id's have been converted to Promises, so clear the queue.
      // nodeIdsToGet = new Set();
      // console.log("idToNextPromises:", idToNextPromises);
      // .reduce((acc, promise, i) => {
      //   acc[nodeIdsToGet[i]] = promise;
      //   return acc;
      // }, {});
      // nodeIdsToGet = [];
      // Add the new promises to our mapping.
      Object.assign(idToPromises, idToNextPromises);
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
    assert.equal(total, 3850);
  })
  .catch(err => {
    console.log("program rejected with err:", err);
  });
