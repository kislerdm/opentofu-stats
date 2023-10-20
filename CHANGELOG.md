## v0.0.2

## Fixed

- Added HTML page generation step to embed the aggregates.json to the HTML file in order to eliminate inconsistency caused
  by the GitHub pages caching.

## v0.0.1

Initial release. It includes the following graphs and metrics:

- _Overall stats_ indicating the current state:
  - Total number of GitHub starts
  - Total numer of tofu downloads
  - Total numer of committers
  - Total numer of recurrent committers this week
  - Total numer of open issues
  - Total numer of open pull requests


- _Timeseries_ indicating the metrics evolution in time, the numbers are broken down by the time frame:
  - Total numer of issues, new and closed.
  - Total numer of pull requests, new and merged.
  - PR time-to-merge averaged over all merged pull requests which were opened throughout a certain time frame.
  - The numer of committers, new, recurrent and total:
    - _New_ committers are defined as the total numer of GitHub users who made their 
      first commit to the [tofu codebase](https://github.com/opentofu/opentofu) throughout a certain time frame.
      For example, a committer is new for the week W41 if their first commit was made on the week W41. 
    - _Recurrent_ committers are defined the total numer of GitHub users who made their
      commit to the [tofu codebase](https://github.com/opentofu/opentofu) throughout a certain time frame and the 
      previous time frame. For example, a committer is recurrent for the week W41, if their also committed during the week W40. 
  - Total numer of commits.

The following time frames are supported:
- Weekly
- Monthly
