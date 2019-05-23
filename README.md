# Drills

Drills is a repository containing upgrade and longevity **exercises** for Concourse. These tests are an extension of the core Concourse testing ([`testflight`](https://github.com/concourse/concourse/tree/master/testflight) and [`topgun`](https://github.com/concourse/concourse/tree/master/topgun)) where the upgrade tests are for testing specific upgrade paths and the longevity tests are for long running pipelines that will test the performance of the current development version.

* **Upgrade path testing**: these are specific, targeted tests for tricky cases we can think of before and after a _specific_ upgrade path like v4.2.2 to v5.0.0.
  
  * Example: the migration for [#2386](https://github.com/concourse/concourse/issues2386) and how we represent paused/disabled resource state.
  * Success means: "before" test suite passes before the upgrade, and the "after" suite passes after the upgrade.
  
* **Longevity/soak testing**: these are not tied to specific upgrade paths, and are meant to be a realistic (yet consistently measurable) workload exercising common _and_ corner cases.
  
  * Example: running unit test builds periodically, exercising things like tagged workers, etc.
  * Success means: the same workloads have the same results before and after the upgrade.
  
* **Performance testing**: similar to longevity testing, but we may have more specific examples (e.g. running Docker builds, running on tagged + untagged workers, running tests that do a bunch of I/O with tiny files, etc. etc.)
  
  * Examples: running Docker builds, running builds with many tasks to exercise container scheduling, configuring pipelines that share resources across teams/etc. to exercise #2386.
  * Success means: resource utilization (e.g. # containers, CPU, RAM) stayed the same or went down, builds took the same amount of time or became faster, and there are no observable leaks over time (goroutine, memory, etc).
  
  
  
## Basic Repository Structure

```
.
├── upgrade                             # Folder that contains the pre and post tests
│   │                                   # for specific upgrade paths. The tests tend to be
│   │                                   # geared more towards focused migration changes.
│   │
│   ├── 4.2.2-5.0.0                     # A specific upgrade path with targeted tests
│   │  	│                               # that we can think of before and after the upgrade
│   │  	│	
│   │   ├── pre                         # The pre ginkgo suite test that contains targeted
│   │   │   │                           # before tests and setup for after upgrade tests
│   │   │   ├── run-all                 # Script that runs all pre ginkgo tests
│   │   │   └── ...                     # Go test files (eg. disabled_versions_tests.go) 
│   │   │
│   │   ├── post                        # The post ginkgo suite test that contains targeted
│   │  	│   │                           # before tests and setup for after upgrade tests
│   │   │   ├── run-all                 # Script that runs all post ginkgo tests
│   │   │   └── ...                     # Go test files (eg. disabled_versions_tests.go) 
│   │   │
│   │   └── pipelines                   # Folder that contains all pipelines used in the pre
│   │                                   # and post tests
│   │
│   └── ...                             # Location to add new upgrade paths
│   
└── longevity                           # Directory that contains all the longevity and
    │                                   # performance tests. This longevity tests contain all
    │                                   # long running pipelines that are set before the upgrade
    │                                   # and observed with the differencesin metrics before and
    │                                   # after the upgrade and with build success/failure
    │ 
    ├── run-all                         # Script that runs the setup script in every pipelines
    │                                   # folder 
    │
    ├── general-pipelines               # This pipelines folder contains pipelines that do not
    │   │                               # require special setup. Meaning that every pipeline in
    │   │                               # this folder is set in one team, with one pipeline
    │   │                               # without any manual assistance to have the pipeline
    │   │                               # run successfully
    │   │
    │   ├── setup                       # Script that setups every pipeline once within the
    │   │                               # same team
    │   └── pipelines                   # Directory to store general pipeline configs
    │       │                           # that needs to be set
    │       │
    │       └── *.yml                   # General pipeline configurations
    │
    ├── shared-check-containers-across-teams  # An example of a longevity test that requires
    │   │                                     # special setup (eg. sets the pipeline 5 times
    │   │                                     # with 5 different teams)
    │   │
    │   ├── setup                       # Setup script that setups the pipeline 5 times in
    │   │                               # 5 different teams
    │   └── pipelines                   # Directory that contains the pipeline configuration
    │       │                           # that needs to be set
    │       │
    │       └── check-containers-stress-test.yml  # The pipeline configuration for the 
    │                                             # longevity test
    └── ...                             # Additional special setup longevity/performance tests
```

## How to run the tests

### Upgrade Tests

  1. Deploy your "before" environment
  
  2. Go into the `pre` folder in the upgrade path you want to run `eg, /drills/upgrade/4.2.2-5.0.0/pre`
  
  3. Set the required environment variables (they should be specified within each folder's README)
  
  4. Run the script to sets up and tests the environment
  ```
  ./run-all
  ````
  
  5. Upgrade your environment
  
  6. Go into the `post` folder `eg, /drills/upgrade/4.2.2-5.0.0/post`
  
  7. Run the script that tests the upgraded environment
  ```
  ./run-all
  ````
  
## Longevity Tests

  1. Deploy your "before" environment
  
  2. Go into the test folder
  ```
    cd /drills/longevity
  ```
  
  3. Run the script that sets all the pipelines
  ```
  ./run-all
  ````
  
  4. Observe the metrics
  
  5. Upgrade your environment
  
  6. Compare the metrics
  
  
### Current Drills environment

* Deployment environment: http://drills.concourse-ci.org/
* Metrics: https://metrics.concourse-ci.org/dashboard/db/concourse?refresh=1m&orgId=1&var-deployment=drills
