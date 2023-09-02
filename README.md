
select date, count, shows_count, click_count, count_uniq_ad_id, count_uniq_campaign_id
from (
    select date,
           count() as count,
           countIf(event = 'show') as 'shows_count',
           countIf(event = 'click') as 'click_count',
           uniqExact(ad_id) as 'count_uniq_ad_id',
           uniqExact(campaign_union_id) as 'count_uniq_campaign_id'
    from ads_data
    group by date
     )

     
select click.ad_id from
                       (select time as time_click,
       ad_id,
       client_union_id,
       campaign_union_id
from t
    prewhere
        event == 'click') as click
inner join
(
    select time as time_show,
                 ad_id,
                 client_union_id,
                 campaign_union_id
          from t
              prewhere event = 'show') as shows
    on
        (click.ad_id == shows.ad_id)
            and (click.client_union_id == shows.client_union_id)
            and (click.campaign_union_id == shows.campaign_union_id)
where time_click > time_show


