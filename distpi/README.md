Distributed Pi
================================================================================

This directory contains an implementation of the [Bailey-Borwein-Plouffe formula][bbp]
and a small wrapper so it can be deployed to [Cloud Functions][gcf].

[bbp]: https://en.wikipedia.org/wiki/Bailey%E2%80%93Borwein%E2%80%93Plouffe_formula
[gcf]: https://cloud.google.com/functions

This is not an official Google product.

Local Development
--------------------------------------------------------------------------------

In order to work with this code you'll need to have a functioning [node.js]
development environment with [yarn] installed. We recommend matching your Node
version to the current recommended Cloud Functions runtime. As of March 2022,
that's Node 16.

[node.js]: https://nodejs.org/en/
[yarn]: https://yarnpkg.com/

To get started, install dependencies and then run the tests:

    $ yarn install
    [ ... ]

    $ yarn test
    PASS  test/modpow.test.js
    PASS  test/bbp.test.js

    [ ... ]

    Test Suites: 2 passed, 2 total
    Tests:       8 passed, 8 total
    Snapshots:   0 total
    Time:        4.512 s
    Ran all test suites.

`index.js` contains a function called `httpCalc`. To test this locally, run
`yarn start` and then open a browser to [localhost:8080](http://localhost:8080).
You can request a different number of digits by adding `?length=<number>` to the
URL. You can also use `offset=<number>` to request digits starting at a different
position.

Deployment
--------------------------------------------------------------------------------

We use [functions-framework] to simplify deployment to [Cloud Functions][gcf].
You'll need to have [gcloud] installed and configured with a project that has
Cloud Functions enabled.

[functions-framework]: https://github.com/GoogleCloudPlatform/functions-framework-nodejs
[gcloud]: https://cloud.google.com/sdk/gcloud

To deploy the sample HTTP function run the following command, replacing
`FUNCTION_NAME` with a name of your choosing:

    $ gcloud functions deploy FUNCTION_NAME \
        --trigger-http \
        --entry-point httpCalc \
        --runtime nodejs16

To verify it is deployed, run `gcloud functions call`:

    $ gcloud functions call FUNCTION_NAME
    executionId: [...]
    result: 3243f6a8885a308d313198a2e03707344a4093822299f31d0082efa98ec4e6c89452821e638d01377be5466cf34e90c6cbf9

The new [second generation Cloud Functions][gcfv2] requires a different invocation:

    $ gcloud beta functions deploy FUNCTION_NAME \
        --gen2 \
        --trigger-http \
        --entry-point httpCalc \
        --runtime nodejs16

    $ gcloud beta functions call --gen2 FUNCTION_NAME
    3243f6a8885a308d313198a2e03707344a4093822299f31d0082efa98ec4e6c89452821e638d01377be5466cf34e90c6cbf9

[gcfv2]: https://cloud.google.com/functions/docs/2nd-gen/overview
