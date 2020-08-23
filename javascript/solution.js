const fetch = require("node-fetch");
const assert = require("assert");

const URL = "https://graph.modulitos.com";

const STARTING_NODE = "a";

// This solution leverages concurrency to calculate the solution in
// the shortest amount of time possible. It calculates the result in 8
// seconds, assuming no additional network latency. To process our
// nodes most efficiently, we will have the following requests in
// flight:

// | time (sec) | nodes requests being processed  |
// |------------+---------------------------------|
// |          1 | a                               |
// |          2 | c, e                            |
// |          3 | f, g, i, j                      |
// |          4 | f, s, t, u, v, w, x             |
// |          5 | f, s, t, u, v, w, x             |
// |          6 | f, s, t, u, v, w, x             |
// |          7 | f, s, t, u, v, w, x             |
// |          8 | s, t, u, v, w, x                |

const main = async () => {
  let nodeIdsToGet = [STARTING_NODE];
  let nodeIdsRetrieved = new Set();
  let idToPromises = {};
  let total = 0;

  while (true) {
    if (nodeIdsToGet.length > 0) {
      const idToNextPromises = nodeIdsToGet.reduce((acc, nodeId) => {
        const promise = (async nodeId => {
          if (nodeIdsRetrieved.has(nodeId)) {
            // Skip processing nodes that have already been fetched.
            return null;
          }
          const res = await fetch(`${URL}/node/${nodeId}`);
          const body = await res.json();
          // Increment the reward:
          total += body.reward;
          // Add the node's children to the queue:
          body.children.forEach(child => nodeIdsToGet.push(child));
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
      // Note: this is not thread safe, but that's okay, this is Node!
      nodeIdsToGet = [];

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

main()
  .then(total => {
    console.log("result:", total);
    assert.equal(total, 4250);
  })
  .catch(err => {
    console.log("program rejected with err:", err);
  });
