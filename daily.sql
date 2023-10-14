WITH
    RECURSIVE dates(date) AS (
        VALUES('2023-08-16')
        UNION ALL
        SELECT date(date, '+1 day')
        FROM dates
        WHERE date >= '2023-08-16' AND date < CURRENT_DATE
    ),
    first_commit AS (SELECT '2023-08-16' AS date),
    commits AS (
        SELECT DATE(author_date) AS date
             , sha
             , raw_author        AS author_id
        FROM main.commits
        WHERE author_date >= (SELECT date FROM first_commit)
    ),
    committers_first_commit_date AS (
        SELECT author_id
            , MIN(date) AS date
        FROM commits
        GROUP BY 1
    ),
    committers_first_commit_cnt AS (
        SELECT date
             , COUNT(1) AS cnt
        FROM committers_first_commit_date
        GROUP BY 1
    ),
    issues_raw AS (
      SELECT * FROM main.issues WHERE type = 'issue'
    ),
    issues_new AS (
        SELECT DATE(created_at)   AS date
             , COUNT(DISTINCT id) AS cnt
        FROM issues_raw
        GROUP BY 1
    ),
    issues_closed AS (
        SELECT DATE(closed_at)    AS date
             , COUNT(DISTINCT id) AS cnt
        FROM issues_raw
        WHERE closed_at IS NOT NULL
        GROUP BY 1
    ),
    issues AS (
        SELECT dates.date
             , SUM(dates.date >= DATE(issues_raw.created_at) AND
                   dates.date < IFNULL(DATE(issues_raw.closed_at), '9999-12-31')) AS cnt_open_total
             , SUM(dates.date >= IFNULL(DATE(issues_raw.closed_at), '9999-12-31')) AS cnt_closed_total
        FROM issues_raw
        CROSS JOIN dates
        GROUP BY 1
    ),
    time_to_merge AS (
        SELECT DATE(created_at)                                AS date_created
             , (UNIXEPOCH(merged_at) - UNIXEPOCH(created_at)) AS time_to_merge_sec
        FROM main.pull_requests
        WHERE merged_at IS NOT NULL
    ),
    time_to_merge_stat AS (
        SELECT *
             , ROW_NUMBER() OVER (PARTITION BY date_created ORDER BY time_to_merge_sec) rn
             , COUNT(1) OVER (PARTITION BY date_created) AS                             cnt
        FROM time_to_merge
    ),
    time_to_merge_median AS (
        SELECT date_created
             , AVG(time_to_merge_sec) / 3600. AS hours
             , cnt
        FROM time_to_merge_stat
        WHERE rn = (cnt / 2) + 1 AND cnt % 2 = 1 OR rn IN (cnt / 2, cnt / 2 + 1) AND cnt % 2 = 0
        GROUP BY 1
    ),
    pr_opened AS (
        SELECT DATE(created_at)    AS date
             , COUNT(DISTINCT id) AS cnt
        FROM main.pull_requests
        WHERE draft IS FALSE
        GROUP BY 1
    ),
    pr_closed AS (
        SELECT DATE(closed_at)    AS date
             , COUNT(DISTINCT id) AS cnt
        FROM main.pull_requests
        GROUP BY 1
    ),
    pr AS (
        SELECT dates.date
             , SUM(dates.date >= DATE(created_at) AND
                   dates.date < IFNULL(DATE(closed_at), '9999-12-31')) AS cnt_open_total
             , SUM(dates.date >= IFNULL(DATE(closed_at), '9999-12-31')) AS cnt_closed_total
        FROM main.pull_requests
        CROSS JOIN dates
        GROUP BY 1
    )

SELECT dates.date
     , IFNULL(count(DISTINCT commits.sha), 0)       AS commits
     , IFNULL(count(DISTINCT commits.author_id), 0) AS committers
     , IFNULL(committers_first_commit_cnt.cnt, 0)   AS committers_new
     , IFNULL(issues_new.cnt, 0)                    AS issues_new
     , IFNULL(issues_closed.cnt, 0)                 AS issues_closed
     , issues.cnt_open_total                        AS issues_open_total
     , issues.cnt_closed_total                      AS issues_closed_total
     , IFNULL(pr_opened.cnt, 0)                     AS pr_new
     , IFNULL(pr_closed.cnt, 0)                     AS pr_closed
     , pr.cnt_open_total                            AS pr_open_total
     , pr.cnt_closed_total                          AS pr_closed_total
     , time_to_merge_median.hours                   AS pr_time_to_merge_median_hours
FROM dates
LEFT JOIN commits ON dates.date = commits.date
LEFT JOIN committers_first_commit_cnt ON dates.date = committers_first_commit_cnt.date
LEFT JOIN issues_new ON dates.date = issues_new.date
LEFT JOIN issues_closed ON dates.date = issues_closed.date
LEFT JOIN issues ON dates.date = issues.date
LEFT JOIN time_to_merge_median ON dates.date = time_to_merge_median.date_created
LEFT JOIN pr_opened ON dates.date = pr_opened.date
LEFT JOIN pr_closed ON dates.date = pr_closed.date
LEFT JOIN pr ON dates.date = pr.date
GROUP BY 1
;
