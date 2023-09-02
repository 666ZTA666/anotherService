
select date, count, shows_count, click_count, count_uniq_ad_id, count_uniq_campaign_id
from (
    select date,
           count() as count,
           countIf(event = 'show') as 'shows_count',
           countIf(event = 'click') as 'click_count',
           uniqExact(ad_id) as 'count_uniq_ad_id',
           uniqExact(campaign_union_id) as 'count_uniq_campaign_id'
    group by date
     )
