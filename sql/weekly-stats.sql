WITH
    RECURSIVE dates_d(date) AS (
        VALUES('2023-08-16')
        UNION ALL
        SELECT date(date, '+1 days')
        FROM dates_d
        WHERE date >= '2023-08-16' AND date < CURRENT_DATE
    ),
    dates_raw AS (
        SELECT DISTINCT STRFTIME('%Y-W%W', date) AS date
        FROM dates_d
    ),
    dates AS (
        SELECT date, row_number() over () AS frame
        FROM dates_raw
    ),
    first_commit AS (SELECT '2023-08-16' AS date),
    commits AS (
        SELECT STRFTIME('%Y-W%W', author_date)  AS date
             , sha
             , raw_author                       AS author_id
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
    committers_recurrent_raw AS (
        SELECT DISTINCT dates.date
                      , dates.frame
                      , commits.author_id
        FROM dates
        LEFT JOIN commits ON dates.date = commits.date
    ),
    committers_recurrent AS (
        SELECT r.date
            , COUNT(1) AS cnt
        FROM committers_recurrent_raw AS l
        INNER JOIN committers_recurrent_raw AS r ON l.author_id = r.author_id AND l.frame+1 = r.frame
        GROUP BY 1
    ),
    issues_raw AS (
      SELECT *
           , STRFTIME('%Y-W%W', created_at) AS date_created
           , STRFTIME('%Y-W%W', closed_at)  AS date_closed
      FROM main.issues WHERE type = 'issue'
    ),
    issues_new AS (
        SELECT date_created       AS date
             , COUNT(DISTINCT id) AS cnt
        FROM issues_raw
        GROUP BY 1
    ),
    issues_closed AS (
        SELECT date_closed        AS date
             , COUNT(DISTINCT id) AS cnt
        FROM issues_raw
        WHERE date_closed IS NOT NULL
        GROUP BY 1
    ),
    issues AS (
        SELECT dates.date
             , SUM(dates.date >= issues_raw.date_created AND
                   dates.date < IFNULL(issues_raw.date_closed, '9999-52'))  AS cnt_open_total
             , SUM(dates.date >= IFNULL(issues_raw.date_closed, '9999-52')) AS cnt_closed_total
        FROM issues_raw
        CROSS JOIN dates
        GROUP BY 1
    ),
    time_to_merge AS (
        SELECT STRFTIME('%Y-W%W', created_at)                              AS date_created
             , AVG((UNIXEPOCH(merged_at) - UNIXEPOCH(created_at))) / 3600. AS time_to_merge_mean_h
             , COUNT(DISTINCT id)                                          AS cnt
        FROM main.pull_requests
        WHERE merged_at IS NOT NULL
        GROUP BY 1
    ),
    pr_opened AS (
        SELECT STRFTIME('%Y-W%W', created_at) AS date
             , COUNT(DISTINCT id)            AS cnt
        FROM main.pull_requests
        WHERE draft IS FALSE
        GROUP BY 1
    ),
    pr_closed AS (
        SELECT STRFTIME('%Y-W%W', closed_at) AS date
             , COUNT(DISTINCT id)           AS cnt
        FROM main.pull_requests
        GROUP BY 1
    ),
    pr AS (
        SELECT dates.date
             , SUM(dates.date >= STRFTIME('%Y-W%W', created_at) AND
                   dates.date < IFNULL(STRFTIME('%Y-W%W', closed_at), '9999-52'))  AS cnt_open_total
             , SUM(dates.date >= IFNULL(STRFTIME('%Y-W%W', closed_at), '9999-52')) AS cnt_closed_total
        FROM main.pull_requests
        CROSS JOIN dates
        GROUP BY 1
    ),
    start_daily AS (
        SELECT dates.date
             , COUNT(DISTINCT user) AS cnt
        FROM dates
        LEFT JOIN main.stars ON dates.date = STRFTIME('%Y-W%W', stars.starred_at)
        INNER JOIN users ON users.id = stars.user
        WHERE users.type = 'User'
        GROUP BY 1
    ),
    stars AS (
        SELECT a.date
             , a.cnt
             , SUM(b.cnt) AS total
        FROM start_daily AS a, start_daily AS b
        WHERE b.date <= a.date
        GROUP BY 1, 2
    ),
    assets AS (
        SELECT STRFTIME('%Y-W%W', fetched_at) AS date
             , asset_name
             , MAX(download_count)            AS cnt
        FROM main.assets
        WHERE asset_name NOT LIKE '%SHA256SUMS%'
        GROUP BY 1, 2
    ),
    downloads AS (
        SELECT date
             , SUM(cnt) AS cnt
        FROM assets
        GROUP BY 1
    )
SELECT dates.date
     , IFNULL(count(DISTINCT commits.sha), 0)       AS commits
     , IFNULL(count(DISTINCT commits.author_id), 0) AS committers
     , IFNULL(committers_first_commit_cnt.cnt, 0)   AS committers_new
     , IFNULL(committers_recurrent.cnt, 0)          AS committers_recurrent
     , IFNULL(issues_new.cnt, 0)                    AS issues_new
     , IFNULL(issues_closed.cnt, 0)                 AS issues_closed
     , issues.cnt_open_total                        AS issues_open_total
     , issues.cnt_closed_total                      AS issues_closed_total
     , IFNULL(pr_opened.cnt, 0)                     AS pr_new
     , IFNULL(pr_closed.cnt, 0)                     AS pr_closed
     , time_to_merge.cnt                            AS pr_merged
     , pr.cnt_open_total                            AS pr_open_total
     , pr.cnt_closed_total                          AS pr_closed_total
     , time_to_merge.time_to_merge_mean_h           AS pr_time_to_merge_mean_hours
     , stars.cnt                                    AS stars_new
     , stars.total                                  AS stars_total
     , downloads.cnt                                AS downloads_total
FROM dates
LEFT JOIN commits ON dates.date = commits.date
LEFT JOIN committers_first_commit_cnt ON dates.date = committers_first_commit_cnt.date
LEFT JOIN issues_new ON dates.date = issues_new.date
LEFT JOIN issues_closed ON dates.date = issues_closed.date
LEFT JOIN issues ON dates.date = issues.date
LEFT JOIN time_to_merge ON dates.date = time_to_merge.date_created
LEFT JOIN pr_opened ON dates.date = pr_opened.date
LEFT JOIN pr_closed ON dates.date = pr_closed.date
LEFT JOIN pr ON dates.date = pr.date
INNER JOIN stars ON dates.date = stars.date
LEFT JOIN committers_recurrent ON dates.date = committers_recurrent.date
LEFT JOIN downloads ON dates.date = downloads.date
GROUP BY 1
;
