
# Async Graph Traversal Problem

One of the trickiest, and yet most common, challenges in programming comes from dealing with concurrency and asynchronous code. We have a server already set up with some API endpoints to query. An example starting point is:

    GET https://graph.modulitos.com/node/a

Each endpoint returns JSON of the form:

    {
       "children": [ "b", "e" ],
       "reward": 1
    }

Where the children in this example can be accessed via:
    GET https://graph.modulitos.com/node/b
    GET https://graph.modulitos.com/node/e


Your challenge is to write an algorithm that traverses the entire collection of nodes and returns the sum of their rewards. Each reward should only be counted a single time. The input to the algorithm will be a URL for a node to begin with, such as https://graph.modulitos.com/node/a.

The node whose JSON result appears above has a reward of 1, and it has links to two other nodes which are part of the collection.

Notes:

 * Some of the endpoints take time to perform their job. You should implement your algorithm in a way that explores the nodes in as short a time as possible.
 * The most important quality of your solution is for it to be correct and bug-free. Optimal and elegant code is appreciated, but those are secondary in importance.
 * The exercise is intended to take less than an evening of work.
 * Test frequently to ensure your design decisions work well.

Please take your time to implement a sound and efficient solution. An optimal solution should take less than ten seconds to process.

## Running the API server locally

Clone the api repo:
https://github.com/modulitos/async-graph-exercise-api/

then run `cargo run`, and the api should be available at `localhost:7878`.

That the api repo link above has more details about how the api works.
