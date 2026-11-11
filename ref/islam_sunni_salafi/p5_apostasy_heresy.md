# Apostasy and heresy: mainstream Sunni and Salafi sources compared

Research notes for the civ book (Principles 4 and 5). Compiled 2026-11-10.

## How to read the verification labels

- **[V-primary]**: I read the passage in the author's or institution's own text and quote it from that text. In most cases this was a digital transcription keyed to the printed page numbers: al-Maktaba al-Shamila (shamela.ws), which labels its texts "ترقيم الكتاب موافق للمطبوع", meaning its pages match the print edition; Arabic Wikisource; or an official website. Where I say "page image", I read a scan of the printed page itself.
- **[V-primary, archived]**: the primary text read through a Wayback Machine snapshot.
- **[secondary]**: a news report or scholarly work that describes or quotes the primary source. I have not seen the primary source.
- **[unverified]**: I could not confirm it. It is included only to mark a gap.
- All translations are mine and are fairly literal. Square brackets mark words I added for sense.
- Local copies of everything I downloaded are under `ext-ref/islam_sunni_salafi/dl_apostasy/` (repo root; not committed).

Access notes. sunnah.com and dorar.net returned Cloudflare challenge pages. For hadith texts I used the fawazahmed0 hadith-api on jsdelivr, which uses the same Fu'ad 'Abd al-Baqi / Fath al-Bari numbering as sunnah.com, and cross-checked against Arabic Wikisource. binbaz.org.sa works only when the full URL with the Arabic slug is used. dar-alifta.org pages load, but its site search does not. Cambridge Core and dawn.com blocked direct download.

---

## Q1. Apostasy (ridda) in classical law

### 1.1 The hadith "man baddala dinahu faqtuluhu"

**Sahih al-Bukhari no. 6922** (Fath al-Bari / sunnah.com numbering), in the book of "Calling Apostates and the Obstinate to Repent, and Fighting Them" (كتاب استتابة المرتدين والمعاندين وقتالهم), chapter "The ruling on the male and female apostate and calling them to repent" (باب حكم المرتد والمرتدة واستتابتهم). [V-primary]

> حَدَّثَنَا أَبُو النُّعْمَانِ، مُحَمَّدُ بْنُ الْفَضْلِ حَدَّثَنَا حَمَّادُ بْنُ زَيْدٍ، عَنْ أَيُّوبَ، عَنْ عِكْرِمَةَ، قَالَ أُتِيَ عَلِيٌّ ـ رضى الله عنه ـ بِزَنَادِقَةٍ فَأَحْرَقَهُمْ فَبَلَغَ ذَلِكَ ابْنَ عَبَّاسٍ فَقَالَ لَوْ كُنْتُ أَنَا لَمْ أُحْرِقْهُمْ لِنَهْىِ رَسُولِ اللَّهِ صلى الله عليه وسلم وَلَقَتَلْتُهُمْ لِقَوْلِ رَسُولِ اللَّهِ صلى الله عليه وسلم " مَنْ بَدَّلَ دِينَهُ فَاقْتُلُوهُ ".

Translation: "'Ikrima said: Some zanadiqa (concealed unbelievers) were brought to 'Ali, and he burned them. This reached Ibn 'Abbas, who said: 'Had it been me, I would not have burned them, because of the prohibition of the Messenger of God; and I would have killed them, because of the saying of the Messenger of God: "Whoever changes his religion, kill him."'"

- Source: hadith-api, ara-bukhari/6922 (https://cdn.jsdelivr.net/gh/fawazahmed0/hadith-api@1/editions/ara-bukhari/6922.json). The same text is on Arabic Wikisource as **no. 6524** in Mustafa al-Bugha's numbering (page "صحيح البخاري/كتاب استتابة المرتدين والمعاندين وقتالهم"). Al-Bugha's 6524 is the same report as Fath al-Bari's 6922.
- The same saying also appears at **Bukhari 3017** (Kitab al-Jihad), where Ibn 'Abbas quotes "Do not punish with God's punishment" (لا تعذبوا بعذاب الله), i.e. fire. [V-primary]
- Bukhari's chapter heading in the same book adds: "وقال ابن عمر والزُهري وإبراهيم: تقتل المرتدة", meaning "Ibn 'Umar, al-Zuhri and Ibrahim [al-Nakha'i] said: the female apostate is killed." [V-primary, Wikisource]
- **Bukhari 6923** (Mu'adh and Abu Musa in Yemen): a man who "was a Jew, became Muslim, then became a Jew again" was killed. Mu'adh refused to sit down "until he is killed, the judgment of God and His Messenger" (لاَ أَجْلِسُ حَتَّى يُقْتَلَ. قَضَاءُ اللَّهِ وَرَسُولِهِ). [V-primary]
- **Bukhari 6878**: "The blood of a Muslim … is lawful only in one of three cases: a life for a life, the married adulterer, and the one who departs from the religion and leaves the community" (وَالْمَارِقُ مِنَ الدِّينِ التَّارِكُ الْجَمَاعَةَ). [V-primary]
- A variant from 'A'isha (**Abu Dawud 4353; al-Nasa'i 4048**) describes the third case as a man "who goes out of Islam waging war against God and His Messenger" (رَجُلٌ يَخْرُجُ مِنَ الإِسْلاَمِ يُحَارِبُ اللَّهَ عَزَّ وَجَلَّ وَرَسُولَهُ), punished by killing, crucifixion or exile (al-Nasa'i's wording). [V-primary] Modern Azhar figures use this variant (see 2b).

### 1.2 Hanafi: al-Marghinani, *al-Hidaya*, "Bab ahkam al-murtaddin"

Edition: ed. Talal Yusuf, Beirut: Dar Ihya' al-Turath al-'Arabi, 1425/2004, vol. 2, pp. 406–407. Shamela book 11820, pages 547–548 (https://shamela.ws/book/11820/547). [V-primary]

> قال: "وإذا ارتد المسلم عن الإسلام والعياذ بالله عرض عليه الإسلام فإن كانت له شبهة كشفت عنه" لأنه عساه اعترته شبهة فتزاح وفيه دفع شره بأحسن الأمرين إلا أن العرض على ما قالوا غير واجب لأن الدعوة بلغته.

"When a Muslim apostatizes from Islam (God protect us), Islam is offered to him; if he has a doubt, it is cleared up for him. [This is] because perhaps a doubt has come over him that can be removed, and this repels his harm by the better of two means. But, as they said, the offering is not obligatory, because the call [to Islam] has already reached him."

> قال: "ويحبس ثلاثة أيام فإن أسلم وإلا قتل ..." وتأويل الأول أنه يستمهل فيمهل ثلاثة أيام ... وعن أبي حنيفة وأبي يوسف رحمهما الله أنه يستحب أن يؤجله ثلاثة أيام طلب ذلك أو لم يطلب

"'He is imprisoned three days; if he becomes Muslim [he is released], otherwise he is killed.' … The first [wording] means that if he asks for time he is given three days … From Abu Hanifa and Abu Yusuf [it is reported] that it is recommended to give him three days, whether or not he asks."

> قال "فإن قتله قاتل قبل عرض الإسلام عليه كره ولا شيء على القاتل" ومعنى الكراهية ههنا ترك المستحب وانتفاء الضمان لأن الكفر مبيح للقتل

"'If someone kills him before Islam is offered to him, that is disliked, and nothing is due from the killer.' 'Disliked' here means omitting what is recommended. There is no liability because unbelief makes killing permissible."

**Women (the Hanafi difference):**

> "وأما المرتدة فلا تقتل" وقال الشافعي ﵀ تقتل لما روينا ... ولنا أن النبي ﵊ نهى عن قتل النساء ولأن الأصل تأخير الجزية إلى دار الآخرة إذ تعجيلها يخل بمعنى الابتلاء وإنما عدل عنه دفعا لشر ناجز وهو الحراب ولا يتوجه ذلك من النساء لعدم صلاحية البنية بخلاف الرجال فصارت المرتدة كالأصلية قال "ولكن تحبس حتى تسلم"

"'As for the female apostate, she is not killed.' Al-Shafi'i said she is killed … Our argument: the Prophet forbade killing women. Also, the principle is that recompense is deferred to the Hereafter, since hastening it would undo the meaning of [this life as] a test. That principle was set aside only to repel an imminent harm, namely warfare (al-hirab). Warfare is not expected from women because of their physical constitution, unlike men, so the female apostate is like the [woman who was] originally [non-Muslim]. He said: 'But she is imprisoned until she becomes Muslim.'"

(The Shamela text reads الجزية; the sense requires الأجزية, "recompenses". I have not checked another edition.)

The next page (p. 407) adds that, according to *al-Jami' al-saghir*, the woman "is compelled to Islam" (تجبر المرأة على الإسلام), and "it is reported that she is beaten at intervals of days, to press her toward Islam" (ويروى تضرب في كل أيام مبالغة في الحمل على الإسلام).

Summary of the Hanafi position:
- The offer of Islam and the three-day respite are recommended, not obligatory.
- A man is killed if he refuses.
- A woman is not killed but is imprisoned and pressured.
- The stated reason for executing men is to prevent "hirab" (warfare).
- A private person who kills an apostate owes nothing.

### 1.3 Maliki: Khalil b. Ishaq, *Mukhtasar Khalil*, "Bab al-ridda"

Edition: ed. Ahmad Jad, Cairo: Dar al-Hadith, 1426/2005, pp. 238–239. Shamela book 11355, pages 307–308 (https://shamela.ws/book/11355/307). The same matn appears in Shamela 784 (*Mawahib al-Jalil min adillat Khalil*, vol. 4, p. 323). [V-primary]

> الردة: كفر المسلم بِصَرِيحٍ أَوْ لَفْظٍ يَقْتَضِيهِ أَوْ فِعْلٍ يَتَضَمَّنُهُ: كإلقاء مصحف بقذر وشد زنار وسحر ... وَاسْتُتِيبَ ثَلَاثَةَ أَيَّامٍ بِلَا جُوعٍ وَعَطَشٍ وَمُعَاقَبَةٍ وَإِنْ لَمْ يَتُبْ فَإِنْ تَابَ وَإِلَّا: قُتِلَ واستبرئت بحيضة

"Apostasy is a Muslim's unbelief by an explicit statement, a word that entails it, or an act that implies it, such as throwing a copy of the Qur'an into filth, wearing the zunnar [belt worn by non-Muslims], or sorcery … He is called to repent for three days, without hunger, thirst or punishment, even if he does not repent [during them]. Then, if he repents [he is left]; otherwise he is killed. And she [a female apostate] is checked for pregnancy by one menstrual cycle [before execution]."

> وَقَتْلُ الْمُسْتَسِرِّ بِلَا اسْتِتَابَةٍ إلَّا أَنْ يَجِيءَ تائبا

"The one who concealed [his unbelief, the zindiq] is killed without being called to repent, unless he comes forward repentant [of his own accord]."

Summary of the Maliki position:
- Istitaba is three days and obligatory, with no starvation or torture.
- Women are executed like men, after a pregnancy check.
- The concealed apostate who is caught is not given istitaba.
- For blasphemy, see Q5.

### 1.4 Shafi'i: al-Nawawi, *Minhaj al-talibin*, and Ibn al-Naqib, *'Umdat al-salik*

**Minhaj**, "Kitab al-ridda". Edition: ed. 'Awad Qasim Ahmad 'Awad, Dar al-Fikr, 1425/2005, p. 293. Shamela 12096, page 325 (https://shamela.ws/book/12096/325). [V-primary]

> هي قطع الإسلام بنية أو قول كفر أو فعل سواء قاله استهزاء أو عنادا أو اعتقادا ... ولا تصح ردة صبي ومجنون ومكره ... وتجب استتابة المرتد والمرتدة وفي قول تستحب وهي في الحال وفي قول ثلاثة أيام فإن أصرا قتلا وإن أسلم صح وترك وقيل: لا يقبل إسلامه إن ارتد إلى كفر خفي كزنادقة وباطنية

"It is cutting off Islam by an intention, a statement of unbelief, or an act, whether said in mockery, obstinacy or belief … Apostasy is not valid from a child, an insane person, or someone under duress … Calling the male and female apostate to repent is obligatory (in one view, recommended), and it is immediate (in one view, three days). If they persist, both are killed. If he becomes Muslim, it is valid and he is left alone. It has been said [a weaker view]: his Islam is not accepted if he apostatized to a hidden unbelief, like the zanadiqa and the Batiniyya."

**'Umdat al-salik**, "Bab al-ridda". Edition: ed. 'Abd Allah b. Ibrahim al-Ansari, Qatar: al-Shu'un al-Diniyya, 1982, p. 233. Shamela 37344, page 228 (https://shamela.ws/book/37344/228). [V-primary]

> منِ ارتَدَّ عنِ الإسلامِ وهو بالغٌ، عاقلٌ، مُختارٌ، استحَقَّ القَتْلَ، ويجِبُ على الإمامِ اسْتِتابَتُهُ، فإنْ رجَعَ إلى الإسلامِ قُبِلَ منْهُ، وإنْ أَبَى قُتِلَ في الحالِ، فإنْ كانَ حُرّاً لمْ يَقْتُلْهُ إلاّ الإمامُ أو نائبُهُ، فإنْ قَتَلَهُ غَيْرُهُ عُزِّر ولا دِيةَ عليهِ، وإنْ كانَ عبْداً فللسَّيِّدِ قتْلُهُ، وإنْ تَكرَّرَتْ رِدَّتُهُ وإسلامُهُ قُبِلَ منْهُ ويُعَزَّرُ.

"Whoever apostatizes from Islam while adult, sane and acting by choice deserves death. The imam must call him to repent. If he returns to Islam, it is accepted from him. If he refuses, he is killed at once. If he is free, only the imam or his deputy may kill him; if someone else kills him, that person is given a discretionary punishment (ta'zir) and owes no blood-money. If he is a slave, his master may kill him. If his apostasy and return to Islam are repeated, his Islam is accepted and he is given ta'zir."

### 1.5 Hanbali: Ibn Qudama, *al-Mughni*, "Kitab al-murtadd"

Edition: Cairo: Maktabat al-Qahira (as labeled by Shamela), vol. 9, pp. 3–9. Shamela 8463, pages 3917–3923 (https://shamela.ws/book/8463/3917). [V-primary]

Consensus claim (p. 3):

> وَأَجْمَعَ أَهْلُ الْعِلْمِ عَلَى وُجُوبِ قَتْلِ الْمُرْتَدِّ.

"The people of knowledge agree by consensus that killing the apostate is obligatory."

Al-Khiraqi's base text that Ibn Qudama comments on (masa'ala 7083):

> وَمَنْ ارْتَدَّ عَنْ الْإِسْلَامِ مِنْ الرِّجَالِ وَالنِّسَاءِ، وَكَانَ بَالِغًا عَاقِلًا، دُعِيَ إلَيْهِ ثَلَاثَةَ أَيَّامٍ، وَضُيِّقَ عَلَيْهِ، فَإِنْ رَجَعَ، وَإِلَّا قُتِلَ.

"Whoever apostatizes from Islam, men and women, being adult and sane, is called back to it for three days and is confined; if he returns [he is left], otherwise he is killed."

Women (fasl 1, p. 3): "There is no difference between men and women in the obligation to kill." Ibn Qudama reports the Hanafi dissent: "Abu Hanifa said: she is compelled to Islam by imprisonment and beating, and is not killed" (وَقَالَ أَبُو حَنِيفَةَ: تُجْبَرُ عَلَى الْإِسْلَامِ بِالْحَبْسِ وَالضَّرْبِ، وَلَا تُقْتَلُ). He rejects it on the basis of the general wording of "man baddala dinahu" and the report of Umm Marwan.

Istitaba (fasl 3, pp. 4–6):

> أَنَّهُ لَا يُقْتَلُ حَتَّى يُسْتَتَابَ ثَلَاثًا. هَذَا قَوْلُ أَكْثَرِ أَهْلِ الْعِلْمِ ... وَرُوِيَ عَنْ أَحْمَدَ، رِوَايَةٌ أُخْرَى، أَنَّهُ لَا تَجِبُ اسْتِتَابَتُهُ، لَكِنْ تُسْتَحَبُّ.

"He is not killed until he has been called to repent [over] three [days]. This is the view of most scholars … Another narration from Ahmad says his istitaba is not obligatory but recommended."

Ibn Qudama cites 'Umar's objection to an execution done without istitaba: "Why did you not imprison him three days, feed him a loaf each day, and call him to repent?" (فَهَلَّا حَبَسْتُمُوهُ ثَلَاثًا، فَأَطْعَمْتُمُوهُ كُلَّ يَوْمٍ رَغِيفًا، وَاسْتَتَبْتُمُوهُ). He also gives a rationale: "apostasy happens only because of a doubt, which does not go away at once, so he must be waited on for a period in which to reflect" (لِأَنَّ الرِّدَّةَ إنَّمَا تَكُونُ لِشُبْهَةٍ ... فَوَجَبَ أَنْ يُنْتَظَرَ مُدَّةً يَرْتَئِي فِيهَا).

Repentance (fasl 5, p. 6): repentance is accepted "whatever the unbelief", in one narration from Ahmad (al-Khallal's choice). In the other narration, "the repentance of the zindiq, and of one whose apostasy is repeated, is not accepted" (لَا تُقْبَلُ تَوْبَةُ الزِّنْدِيقِ، وَمَنْ تَكَرَّرَتْ رِدَّتُهُ), which is also the view of Malik.

Who may execute (fasl 7089, pp. 8–9):

> وَقَتْلُ الْمُرْتَدِّ إلَى الْإِمَامِ، حُرًّا كَانَ أَوْ عَبْدًا. وَهَذَا قَوْلُ عَامَّةِ أَهْلِ الْعِلْمِ ... فَإِنْ قَتَلَهُ غَيْرُ الْإِمَامِ، أَسَاءَ، وَلَا ضَمَانَ عَلَيْهِ؛ لِأَنَّهُ مَحَلٌّ غَيْرُ مَعْصُومٍ ... وَعَلَى مَنْ فَعَلَ ذَلِكَ التَّعْزِيرُ؛ لِإِسَاءَتِهِ وَافْتِيَاتِهِ.

"Killing the apostate belongs to the imam, whether he is free or a slave. This is the view of nearly all scholars … If someone other than the imam kills him, he has done wrong, but owes no compensation, because [the apostate] is not a protected person … The one who did it receives ta'zir for his wrongdoing and for usurping [the imam's authority]."

A caveat on p. 11: an apostate who joins enemy territory "may be killed by anyone, without istitaba" (يُبَاحُ قَتْلُهُ لِكُلِّ أَحَدٍ مِنْ غَيْرِ اسْتِتَابَةٍ).

### 1.6 Comparison of the four manuals [all V-primary, above]

| | Hanafi (*Hidaya*) | Maliki (*Khalil*) | Shafi'i (*Minhaj*, *'Umda*) | Hanbali (*Mughni*) |
|---|---|---|---|---|
| Penalty for adult sane male | death | death | death | death |
| Istitaba | recommended, 3 days if requested | obligatory, 3 days, no starvation or torture | obligatory, immediate (minority view: 3 days) | obligatory, 3 days (one narration: recommended) |
| Female apostate | not killed; imprisoned (and beaten, per one report) until she returns | killed, after a pregnancy check | killed | killed |
| Concealed apostate (zindiq) | two narrations | killed without istitaba unless he comes forward repentant | repentance accepted (weaker view: not accepted from zanadiqa or Batiniyya) | two narrations |
| Who executes | not addressed in the passages read; a private killer owes nothing | not addressed in the passage read | imam or deputy only; a private killer gets ta'zir, no diya | imam only; a private killer "did wrong", gets ta'zir, no compensation |

**Salafi relation to this material.** The Salafi tradition does not have a separate classical law of apostasy. Ibn Baz cites "the chapter on the apostate in all four schools" (2a below), and the Saudi establishment follows the Hanbali school in fiqh (general background, not sourced in this file). On the classical *law* of apostasy, the mainstream Sunni and Salafi sources are the same texts. The differences arise in (1) modern reinterpretation (Q2) and (2) what counts as an act of apostasy, i.e. takfir (Q3–Q4).

---

## Q2. Apostasy today

### 2a. Saudi / Salafi

**Ibn Baz, *Majmu' fatawa wa-maqalat*, vol. 9, p. 303.** Answer dictated for *al-Majalla al-'Arabiyya*, dated 3/6/1416 AH (about 27–28 October 1995 by tabular conversion). Online: https://binbaz.org.sa/fatwas/2518 (Arabic slug required). [V-primary]

> قد دل القرآن الكريم والسنة المطهرة على قتل المرتد إذا لم يتب ... وفي صحيح البخاري عن ابن عباس رضي الله عنهما عن النبي ﷺ أنه قال: من بدل دينه فاقتلوه ... والأدلة في هذا كثيرة، وقد أوضحها أهل العلم في باب حكم المرتد في جميع المذاهب الأربعة ... فمن أنكر ذلك فهو جاهل أو ضال لا يجوز الالتفات إلى قوله، بل يجب أن ينصح ويعلم لعله يهتدي.

"The Noble Qur'an and the pure Sunna indicate that the apostate is killed if he does not repent … In Sahih al-Bukhari, from Ibn 'Abbas, the Prophet said: 'Whoever changes his religion, kill him' … The evidence on this is extensive, and the scholars have set it out in the chapter on the ruling of the apostate in all four schools … Whoever denies this is ignorant or astray; his view must not be given attention. Rather, he must be advised and taught, so that perhaps he will be guided."

Two short oral answers by Ibn Baz on the same site:
- https://binbaz.org.sa/fatwas/22451: "مَن بدَّل دينَه يُسْتَتاب، فإن تابَ وإلا قُتِلَ", meaning "Whoever changes his religion is called to repent; if he repents [he is left], otherwise he is killed." [V-primary]
- https://binbaz.org.sa/fatwas/6445 (from his lessons on *Bulugh al-maram*): the wisdom of killing apostates lies in two things, "جزاء له على ما صنع، وعدم توبته. وردعًا لغيره", meaning "recompense for what he did and for not repenting, and deterrence of others." [V-primary]

**Permanent Committee for Scholarly Research and Ifta', fatwa no. 21166, questions 1–3.** *Fatawa al-Lajna al-Da'ima*, first collection, compiled by Ahmad al-Duwaysh, Riyadh: Ri'asat Idarat al-Buhuth al-'Ilmiyya wa-l-Ifta', **vol. 22, pp. 231–236** (Shamela 8381, pages 16856–16863: https://shamela.ws/book/8381/16857). Signed by Bakr Abu Zayd, Salih al-Fawzan and 'Abd Allah Ghudayyan (members) and 'Abd al-'Aziz Al al-Shaykh (chairman). [V-primary]

Q1 (p. 231), the forms of apostasy:

> فالردة بالقول: كسب الله تعالى، أو سب رسوله صلى الله عليه وسلم. والردة بالفعل: كالسجود للصليب أو الصنم، أو الذبح للقبور، أو امتهان المصحف. والردة بالشك ... والردة بالترك: كالإعراض عن دين الإسلام؛ لا يتعلمه ولا يعمل به

"Apostasy by word: such as reviling God or His Messenger. By act: such as prostrating to a cross or an idol, slaughtering [animals] for graves, or desecrating the Qur'an. By doubt … By omission: such as turning away from the religion of Islam, neither learning it nor practicing it."

Q2 (pp. 231–234):

> عقوبة المرتد عن دين الإسلام هي القتل ... وهذا التشديد في عقوبة المرتد لأمور عديدة منها: ١ - أن هذه العقوبة زجر لمن يريد الدخول في الإسلام مصانعة أو نفاقا ... ٢ - من أعلن إسلامه فقد دخل في جماعة المسلمين ... والردة عن الإسلام خروج على جماعة المسلمين ... ٣ - أن المرتد قد يرى فيه ضعفاء الإيمان ... أنه ما ترك الإسلام إلا عن معرفة بحقيقته ... فقتل هذا المرتد إذا هو الواجب؛ حماية للدين الحق من تشويه الأفاكين

"The punishment of the apostate from Islam is death … This severity has several reasons, among them: (1) the punishment deters anyone who would enter Islam insincerely or hypocritically … (2) whoever announces his Islam has entered the community of Muslims … and apostasy is a revolt against the community of Muslims … (3) the weak in faith may think the apostate left Islam only from knowing its reality … so killing this apostate is obligatory, to protect the true religion from the distortions of liars."

Q3 (pp. 234–236) addresses why someone who *enters* Islam is not punished in the same way. The answer: entering Islam is obedience to God, while leaving it is disobedience that "deserves the punishment of this world by death" (فاستحق بذلك عقوبة الدنيا بالقتل).

**Fatwa no. 4792, q. 4 (vol. 22, pp. 236–237)**, signed with Ibn Baz as chairman: a person who apostatized in the past, in a country that does not apply sharia, and who has since repented, is **not** to be punished:

> من ارتد عن دين الإسلام ورجع إليه تائبا نادما فلا يجوز أن يقام عليه الحد؛ لأن الحد يقام على المصر المستمر على ردته

"Whoever apostatized from Islam and returned to it repentant and regretful must not have the hadd applied to him, because the hadd is applied to one who persists in his apostasy." [V-primary]

**IslamQA no. 12406** (dated 7 Safar 1425 / 28 March 2004; https://islamqa.info/ar/answers/12406) reproduces the text of fatwa 21166 Q2 word for word. It cites it as "فتاوى اللجنة الدائمة (21/231-234)". The volume number differs from the Shamela edition (vol. 22), probably because the two use different printings or volume numbering. [V-primary]

### 2b. Al-Azhar and Egypt's Dar al-Ifta

**Ali Gomaa (then Grand Mufti of Egypt), Washington Post–Newsweek "On Faith" site, section "Muslims Speak Out", entry "Sheikh Ali Gomah", segment titled "Freedom of Religion in Islam", posted July 21, 2007.** Read from the Wayback snapshot of 11 Aug 2007: http://web.archive.org/web/20070811123844/http://newsweek.washingtonpost.com:80/onfaith/muslims_speak_out/2007/07/sheikh_ali_gomah.html [V-primary, archived; English original]

> "The essential question before us is can a person who is Muslim choose a religion other than Islam? The answer is yes, they can, because the Quran says, 'Unto you your religion, and unto me my religion,' [Quran, 109:6] … But from a religious perspective, the act of abandoning one's religion is a sin punishable by God on the Day of Judgment. If the case in question is one of merely rejecting faith, then there is no worldly punishment. If, however, the crime of undermining the foundations of the society is added to the sin of apostasy, then the case must be referred to a judicial system whose role is to protect the integrity of the society. … In some cases, this sin of the individual may also represent a greater break with the commonly held values of a society in an attempt to undermine its foundations or even attack its citizenry. Depending on the circumstances, this may reach the level of a crime of sedition against one's society. … This was the basis of the Islamic perspective on apostasy when committed at certain times and under certain circumstances."

After press controversy in Egypt, Gulf News quoted Gomaa as saying: "What I actually said is that Islam prohibits a Muslim from changing his religion and that apostasy is a crime, which must be punished." [secondary: Gulf News, July 2007, as quoted on a blog that reproduces the story; I did not find Gomaa's own Arabic statement or a Dar al-Ifta release]. The 2007 text itself makes worldly punishment conditional on "undermining the foundations of the society".

**Ahmed al-Tayeb (Shaykh al-Azhar), Ramadan television program "al-Imam al-Tayyib", episode broadcast 16 June 2016.** As reported by Parlmany (Youm7 group), 17 June 2016 (https://www.parlmany.com/News/12/94869/), and Elwatan News, 16 June 2016 (https://www.elwatannews.com/news/details/1229224). [secondary: news reports quoting the broadcast and an Azhar summary; I did not locate the text on azhar.eg]

> الردة المعاصرة تظهر فى ثوب الجريمة والاعتداء والخيانة العظمى، ونتعامل معها الآن، على أنها جريمة يجب أن تقاوم وأن تكون فيها عقوبة من العقوبات التعزيرية ... جمهور الفقهاء وأئمة المذاهب الأربعة يعتبرون الردة جريمة، ويتفقون على أن المرتد يستتاب فى مدة مختلف فيها وإلا يقتل ... إلا الأحناف فقالوا: المرأة المرتدة لا تقتل؛ لأنها لا يتصور منها الخروج على المجتمع، وفى هذا ما يؤكد لنا أن الردة أو القتل كحد للردة مشروط بأن يكون المرتد خطر على المجتمع ... وقد ذهب الفقهاء المعاصرون من أمثال شيخنا أبو زهرة والشيخ شلتوت والشيخ عبد الوهاب خلاف وغيرهم، إلى أن الردة جريمة ولكن ليس فيها حد محدد، وإنما تترك لتقدير الحاكم ... وفيها تعزير ... والتعزير فيه مرونة حيث يبدأ باللوم والعتاب والضرب للتأديب وينتهى بالقتل ... واجبنا نحو المرتدين هو النصح والإرشاد ومخاطبة عقولهم بالحجة والبرهان.

"Contemporary apostasy appears in the garb of crime, aggression and high treason, and we now treat it as a crime that must be resisted and that carries one of the discretionary (ta'zir) punishments … The majority of jurists and the imams of the four schools consider apostasy a crime and agree that the apostate is called to repent for a disputed period, otherwise he is killed … except the Hanafis, who said the female apostate is not killed because revolt against society is not conceivable from her. This confirms to us that apostasy, or death as the hadd for apostasy, is conditional on the apostate being a danger to society … Contemporary jurists such as our shaykh Abu Zahra, Shaykh Shaltut, Shaykh 'Abd al-Wahhab Khallaf and others held that apostasy is a crime but has no fixed hadd; it is left to the ruler's assessment … and carries ta'zir … Ta'zir is flexible: it begins with blame and reproach and beating for discipline, and ends with killing … Our duty toward apostates is advice and guidance and addressing their minds with argument and proof."

The same report says he described an apostate "who is a caller to his apostasy" as committing "in my belief, high treason and a revolt against society and its sacred things" (فهذه فى اعتقادى خيانة عظمى وخروج على المجتمع ومقدساته). Elwatan quotes him: "freedom of belief is one thing and freedom of apostasy is another" (حرية الاعتقاد شيء وحرية الارتداد شيء آخر), and says it is "unjust" to judge the question by a human-rights framework that "arose in an environment radically different".

**Official Dar al-Ifta fatwa on the apostate's punishment: not located.** Dar al-Ifta's site search did not work for me, and web searches returned only secondary reports. [unverified]

### 2c. Lay opinion: Pew Research Center, *The World's Muslims: Religion, Politics and Society*, 30 April 2013

PDF: https://www.pewresearch.org/wp-content/uploads/sites/7/2013/04/worlds-muslims-religion-politics-society-full-report.pdf [V-primary]

Fieldwork: surveys conducted 2008–2012 in 39 countries and territories, with over 38,000 face-to-face interviews (Preface). **Saudi Arabia was not surveyed** (nor China, India, Syria).

Question Q92b: "Do you favor or oppose the following: the death penalty for people who leave the Muslim religion?"

**Population 1: Muslims who favor making sharia the official law of the land (Q79a = favor).** Chart "Death Penalty for Leaving Islam", report p. 55 (printed page number). Exact chart subtitle: "Among Muslims who say sharia should be the law of the land, % who favor the death penalty for converts". Figures:

| Country | % |
|---|---|
| Egypt | 86 |
| Jordan | 82 |
| Afghanistan | 79 |
| Pakistan | 76 |
| Palestinian territories | 66 |
| Malaysia | 62 |
| Lebanon | 46 |
| Bangladesh | 44 |
| Iraq | 42 |
| Tunisia | 29 |
| Thailand | 27 |
| Tajikistan | 22 |
| Indonesia | 18 |
| Turkey | 17 |
| Bosnia | 15 |
| Russia | 15 |
| Kyrgyzstan | 14 |
| Kosovo | 11 |
| Albania | 8 |
| Kazakhstan | 4 |

Notes to the chart:
- Thailand: five southern provinces only.
- Russia and Thailand: the sharia question was asked about Muslim areas only.
- Azerbaijan: not shown because the sample was too small.
- The text says: "in six of the 20 countries where there are adequate samples for analysis, at least half of those who favor making Islamic law the official law also support executing apostates."
- Regional medians on p. 23 ("What Do Sharia Supporters Want?"): South Asia 76, Middle East–North Africa 56, Southeast Asia 27, Central Asia 16, Southern–Eastern Europe 13.

**Population 2: all Muslim respondents.** Topline, Q92b, p. 219. The topline marks filtered questions with "ASK IF …" (for example Q81, "ASK IF RESPONDENT SAYS FAVOR IN Q79a"). Q92b has no such marker, so these figures are for all Muslims surveyed. Favor percentages:

| Country | % |
|---|---|
| Egypt | 88 |
| Jordan | 83 |
| Afghanistan | 79 |
| Pakistan | 75 |
| Palestinian territories | 62 |
| Djibouti | 62 |
| Malaysia | 58 |
| DR Congo | 44 |
| Bangladesh | 43 |
| Iraq | 41 |
| Thailand | 23 |
| Tunisia | 18 |
| Lebanon | 17 |
| Indonesia | 16 |
| Russia | 12 |
| Kyrgyzstan | 9 |
| Turkey | 8 |
| Tajikistan | 8 |
| Bosnia | 4 |
| Kosovo | 3 |
| Albania | 2 |
| Kazakhstan | 1 |
| Azerbaijan | 1 |

Most sub-Saharan figures come from Pew's 2008–09 "Tolerance and Tension" survey: Niger 32, Nigeria 29, Senegal 35, Kenya 32 and others in the 19–44 range.

For reference, support for making sharia the official law (Q79a), report p. 46: Afghanistan 99, Iraq 91, Palestinian territories 89, Malaysia 86, Pakistan 84, Morocco 83, Bangladesh 82, Egypt 74, Indonesia 72, Jordan 71, Tunisia 56, Lebanon 29, Turkey 12.

So the widely quoted "86% in Egypt" refers to Egyptian Muslims who favor sharia as official law. The share of *all* Egyptian Muslims who favored the death penalty for leaving Islam was 88% (topline).

---

## Q3. Heresy and takfir, historical

### 3a. Al-Ghazali, *Fada'ih al-Batiniyya* (al-Mustazhiri), Chapter 8: "On the ruling of the law concerning them: declaring them unbelievers and shedding their blood"

Edition: ed. 'Abd al-Rahman Badawi, Kuwait: Mu'assasat Dar al-Kutub al-Thaqafiyya, pp. 146–163. Shamela 6554, pages 146–163 (https://shamela.ws/book/6554/146). [V-primary]

Procedure before takfir (p. 146):

> ومهما سئلنا عن واحد منهم أو عن جماعتهم وقيل لنا هل تحكمون بكفرهم لم نتسارع إلى التكفير إلا بعد السؤال عن معتقدهم ومقالتهم ونراجع المحكوم عليه أو نكشف عن معتقدهم بقول عدول يجوز الاعتماد على شهادتهم فإذا عرفنا حقيقة الحال حكمنا بموجبه

"Whenever we are asked about one of them, or about their group, and told 'Do you judge them unbelievers?', we do not hasten to takfir except after asking about their belief and doctrine. We question the person being judged, or we establish their belief through the statement of upright witnesses whose testimony may be relied on. When we know the true state of affairs, we judge accordingly."

Two grades. First grade, error and innovation, not unbelief (pp. 146–147): a commoner who believes the imamate belongs to the Prophet's family, that 'Ali was wrongly deprived of it, and that the imam is infallible, "and yet does not consider our blood lawful and does not believe us unbelievers":

> فهذا الشخص لا يستباح سفك دمه ولا يحكم بكفره لهذه الأقاويل بل يحكم بكونه ضالا مبتدعا فيزجر عن ضلاله وبدعته بما يقتضيه رأي الإمام

"This person's blood is not made lawful, and he is not judged an unbeliever for these statements. He is judged astray and an innovator, and is restrained from his error and innovation by whatever the imam judges appropriate."

Al-Ghazali also says that a man who holds Abu Bakr and 'Umar to have been sinful (fasiq) is not an unbeliever (p. 149). On calling a Muslim an unbeliever (p. 150): "thinking a Muslim is an unbeliever is not [itself] unbelief" (وظن الكفر بمسلم ليس بكفر), unless the person calls the true doctrines themselves unbelief.

Second grade, unbelief (p. 151):

> المرتبة الثانية المقالات الموجبة للتكفير وهي أن يعتقد ما ذكرناه ويزيد عليه فيعتقد كفرنا واستباحة أموالنا وسفك دمائنا فهذا يوجب التكفير لا محالة

"The second grade, the doctrines that require takfir: that he believes what we mentioned and in addition believes us to be unbelievers and our property and blood to be lawful. This requires takfir without doubt." Denying bodily resurrection, Paradise and Hell, or affirming two gods, also requires takfir (pp. 151–154).

Legal status of those judged unbelievers (Fasl 2, p. 156):

> والقول الوجيز فيه أنه يسلك مسلك المرتدين في النظر في الدم والمال والنكاح والذبيحة ونفوذ الأقضية وقضاء العبادات ... وإنما الواجب قتلهم وتطهير وجه الأرض منهم هذا حكم الذين يحكم بكفرهم من الباطنية وليس يختص جواز قتلهم ولا وجوبه بحالة قتالهم بل نغتالهم ونسفك دماءهم

"In brief: they are treated as apostates with respect to blood, property, marriage, slaughtered meat, the validity of judgments, and making up acts of worship … [The imam has no option of enslaving them, taking jizya, releasing or ransoming.] What is required is to kill them and cleanse the face of the earth of them. This is the ruling for those Batinis who are judged unbelievers. The permission and obligation to kill them is not limited to when they are fighting; rather we assassinate them and shed their blood."

Women (p. 156): "As for the women, we kill them whenever they openly profess the belief that is unbelief … for the female apostate is killed in our view by the generality of 'whoever changes his religion, kill him'. Yes, the imam may follow his own ijtihad: if he chooses the way of Abu Hanifa and refrains from killing the women, the question is one open to ijtihad."

Repentance (Fasl 3, pp. 160–163): al-Ghazali reports the disagreement over accepting a zindiq's repentance and then gives his own view in three cases:
1. One who repents voluntarily, without fear: "his repentance should certainly be accepted".
2. A commoner who becomes Muslim "under the shadow of swords" but is not one of their missionaries: "his repentance is also accepted" (فهذا أيضا تقبل توبته).
3. A missionary (da'i): "the better [view] is that the imam is neither obliged to kill him nor forbidden to; it is left to his ijtihad" (والأولى ألا يوجب على الإمام قتله لا محالة ولا أن يحرم قتله بل يفوض إلى اجتهاده).

On case 2 he reasons (p. 163) that wrongly killing someone who is outwardly Muslim and may be inwardly Muslim "is forbidden", while overlooking a concealed unbeliever who has no spreading mission "is not a great prohibited thing".

### 3b. The Ottoman Shaykh al-Islam Ebussuud Efendi (d. 1574) on the Kızılbaş

Source: M. Ertuğrul Düzdağ, *Şeyhülislâm Ebussuud Efendi Fetvaları: Kanuni Devrinde Osmanlı Hayatı*, İstanbul: Kapı Yayınları, 1st printing 2012 (first published 1972), Part "III. Mürtedler, A. Kızılbaşlar", fatwas **nos. 479–487, pp. 135–139**. Düzdağ transliterates the Ottoman Turkish into modern Turkish script from two manuscripts: (A) Fatih, Ali Emiri, Şer'iye 80, copied 1041/1634; and (B) Bayezid 2757. Folio references are given in brackets. Read from the archive.org PDF text layer (https://archive.org/details/m.-ertugrul-duzdag-seyhulislam-ebussud-efendi-fetvalari-kanuni-devrinde-osmanli-). [V-primary in transliteration; I did not see the Ottoman-script manuscripts]

No. 479 (A. 255b):

> Mes'ele: Kızılbaş tâifesinin şer'an kıtâli helâl olup, katl eden gâzî ve kızılbaş tâifesinin ellerinde maktul olanlar şehîd olurlar mı?
> Elcevap: Olur, gazâ-i ekber ve şehâdet-i azîmedir.
> Su'al-i âhar: Kıtâlleri helâl olduğu takdirce, mahzâ sultân-ı ehl-i İslâm hazretlerine bağy ve adâvet üzere olup, asker-i İslâm'a kılıç çektiği için mi olur, yâhud gayri sebebi var mıdır?
> Elcevap: Hem bâgîlerdir, hem vücûh-i kesîreden kâfirlerdir.

"Question: Is fighting the Kızılbaş lawful under the sharia, and are those who kill them ghazis and those killed by them martyrs? Answer: Yes. It is the greatest holy war and a great martyrdom. Further question: If fighting them is lawful, is it only because they rebel against and show enmity to the sultan of the Muslims and drew the sword against the army of Islam, or is there another reason? Answer: They are both rebels (bâgî) and, on many grounds, unbelievers."

No. 481 (A. 256a). An Arabic subscription found in another manuscript (Köprülü K II 362) dates it to 955 AH (1548–49). The grounds of unbelief listed:
- contempt for the Qur'an and the sharia;
- burning books of the sharia;
- insulting scholars;
- prostrating to their leader;
- declaring lawful what is definitively forbidden;
- cursing Abu Bakr and 'Umar;
- slandering 'A'isha, which amounts to denying the Qur'an and insulting the Prophet.

Then:

> … cumhûr-i ulemâ-i a'sâr ve emsâr icmâı ile, katilleri mubâh olup, küfürlerinde şek edenler kâfir olurlar. İmâm-ı A'zam ve İmam Süfyân-i Sevrî ve İmam Evzâ'î (rahimehumullah) katlarında tamam sıhhat üzere tevbe edip İslâm'a gelicek, eğerçi bu küfürler dahi sâir kefere küfürleri gibi afv olunup katilden necât bulurlar, amma İmam Mâlik ve İmam Şâfi'î ve İmam Ahmed bin Hanbel ve İmam Leys bin Sa'd ve İmam İshak bin Râhûye ve sâir uzemâ-i ulemâ-i dinden cem'-i kesîr katlarında asla tövbeleri makbûl ve İslâm'ları mu'teber değildir. Elbette hadden katl olunurlar. Hazret-i İmâm-ı Din-penah … zikr olunan eimme-i dinden, hangi cânibin kavli ile amel ederler ise meşrû'dur.

"… by the consensus of the majority of scholars of all ages and lands, killing them is permitted, and those who doubt their unbelief become unbelievers. In the view of Abu Hanifa, Sufyan al-Thawri and al-Awza'i, if they repent soundly and come to Islam, these unbeliefs are pardoned like those of other unbelievers and they escape death. But in the view of Malik, al-Shafi'i, Ahmad b. Hanbal, al-Layth b. Sa'd, Ishaq b. Rahuya and many other great scholars, their repentance is never accepted and their Islam is not valid; they are certainly killed as a hadd. Whichever of these imams' views the Imam, Protector of the Faith, acts on is lawful."

The limiting clause in the same fatwa:

> Askerlerinden olup kıtâle mübâşeret edenler ve binip inip etbâ'ından olanların şânında aslâ tevakkuf olunur değildir. Amma şehirlerde ve köylerde kendi hâlinde salâh üzerine olup, bunların sıfatlarından ve ef'âllerinden tenezzühü olup, zâhir halleri dahi sıdklarına delâlet eyleyen kimselerin kizbleri zâhir olmayınca, üzerlerine bunların ahkâmı ve ukûbâtı icrâ olunmaz.

"There is no hesitation regarding those who are among their soldiers and engage in fighting, and those of their followers who ride with them. But those who live peaceably in towns and villages, keeping to themselves in good conduct, free of these people's characteristics and acts, and whose outward state indicates their truthfulness: as long as their lying has not become evident, these rulings and punishments are not applied to them."

Further fatwas:
- No. 482: Kızılbaş children captured in the Nakhchivan campaign do not become slaves.
- No. 483: Armenians captured with the Kızılbaş are to be freed if they did not fight.
- Nos. 484–485: enslaving Kızılbaş women is permissible on a narration attributed to Abu Hanifa, when it strengthens the army of Islam. Their service is lawful, but sexual intercourse is not until they become Muslim, "for they are apostates" (mürteddelerdir).
- No. 486 (B. 302b): if a private person kills a man recorded in the court register as a Kızılbaş who cursed the four caliphs, "if it is established that he killed him at the time he cursed, no action is taken against him" (Sebb ettiği vakit katl ettiği muhakkak ise ta'arruz olunmaz).

Note a discrepancy. Ebussuud attributes non-acceptance of repentance to al-Shafi'i, but al-Nawawi's *Minhaj* (1.4 above) gives acceptance as the relied-on Shafi'i view and non-acceptance only as a weaker view ("وقيل"). Al-Mughni (fasl 5) also lists al-Shafi'i among those who accept.

### 3c. Muhammad b. 'Abd al-Wahhab

**Nawaqid al-Islam (the ten nullifiers).** Arabic Wikisource, page "نواقض الإسلام" (https://ar.wikisource.org/wiki/نواقض_الإسلام). Also checked against Ibn Baz's edition (Saudi Ministry of Islamic Affairs, 1410; Shamela 30869). Ibn Baz's version paraphrases nullifier 1 as "calling on the dead and seeking their help, and vowing and sacrificing to them" and adds his own commentary on man-made law. [V-primary]

> الأَوَّلُ: الشِّرْكُ فِي عِبَادَةِ اللهِ ... وَمِنْهُ الذَّبْحُ لِغَيْرِ اللهِ، كَمَنْ يَذْبَحُ لِلْجِنِّ أَوْ لِلْقَبْرِ.
> الثَّانِي: مَنْ جَعَلَ بَيْنَهُ وَبَيْنَ اللهِ وَسَائِطَ يَدْعُوهُمْ وَيسْأَلُهُمْ الشَّفَاعَةَ، وَيَتَوَكَّلُ عَلَيْهِمْ كَفَرَ إِجْمَاعًا.
> الثَّالِثُ: مَنْ لَمْ يُكَفِّرْ المُشْرِكِينَ أَوْ شَكَّ فِي كُفْرِهِمْ، أَوْ صَحَّحَ مَذْهَبَهُم،ْ كَفَرَ.
> ...
> وَلَا فَرْقَ فِي جَمِيعِ هَذِهِ النَّوَاقِضِ بَيْنَ الهَازِلِ وَالجَادِّ وَالخَائِفِ إِلَّا المُكْرَهِ.

"First: associating others with God in worship … including sacrificing to other than God, such as one who sacrifices to the jinn or to a grave. Second: whoever places intermediaries between himself and God, calling on them, asking them for intercession and relying on them, is an unbeliever by consensus. Third: whoever does not declare the polytheists unbelievers, or doubts their unbelief, or considers their way correct, is an unbeliever. … In all these nullifiers there is no difference between the joker, the serious person and the fearful, except the one under compulsion."

The other nullifiers:
- 4: holding another guidance or ruling better than the Prophet's;
- 5: hating anything the Prophet brought;
- 6: mocking the religion;
- 7: sorcery;
- 8: aiding polytheists against Muslims;
- 9: believing some people may go outside Muhammad's law, as al-Khidr did with Moses;
- 10: turning away from the religion, not learning or practicing it.

**Kashf al-shubuhat.** Arabic Wikisource, page "كشف الشبهات". The Saudi Awqaf edition (1418) is on Shamela as book 30850. [V-primary]

> وعرفت أن إقرارهم بتوحيد الربوبية لم يدخلهم في الإسلام، وأن قصدهم الملائكة أو الأنبياء، أو الأولياء، يريدون شفاعتهم، والتقرب إلى الله بذلك هو الذي أحل دماءهم وأموالهم

"And you know that their affirming God's lordship did not bring them into Islam, and that their turning to the angels, prophets or saints, seeking their intercession and closeness to God through that, is what made their blood and property lawful." (This is about the polytheists the Prophet fought; the book then applies the same reasoning to people "of our time".)

> فاعلم أن شرك الأولين أخف من شرك أهل زماننا بأمرين

"Know that the shirk of the earlier [polytheists] was lighter than the shirk of the people of our time, in two respects." (1) The earlier polytheists called on others only in times of ease and turned to God alone in hardship. (2) They called on righteous beings, while "the people of our time" call on "some of the most dissolute people".

On graves specifically: "we objected to the istighatha of worship that they perform at the graves of the saints or in their absence, in things only God can do" (ونحن أنكرنا استغاثة العبادة التي يفعلونها عند قبور الأولياء أو في غيبتهم في الأشياء التي لا يقدر عليها إلا الله).

Limits stated in the same text:
- On the hadith of Usama: "whoever outwardly shows tawhid and Islam must be left alone unless something contradicting that becomes clear from him" (إن من أظهر التوحيد والإسلام وجب الكف عنه إلا أن يتبين منه ما يناقض ذلك).
- On the Dhat Anwat story: "a Muslim who speaks words of unbelief without knowing it, and is alerted and repents at once, does not become an unbeliever" (أن المسلم إذا تكلم بكلام كُفر وهو لا يدري فنبه على ذلك فتاب من ساعته، أنه لا يكفر).

**His own limiting statement: *al-Durar al-saniyya*, ed. Ibn Qasim, 6th ed., 1417/1996, vol. 1, pp. 102–104**, "Answer of Shaykh Ibn 'Abd al-Wahhab on what is fought over and what makes one an unbeliever" (Shamela 3055, pages 96–98: https://shamela.ws/book/3055/98). [V-primary]

> ولا نكفر إلا ما أجمع عليه العلماء كلهم، وهو: الشهادتان. وأيضا: نكفره بعد التعريف إذا عرف وأنكر.

"We declare unbelief only for what all the scholars agree on, namely [failure to make] the two testimonies. Also, we declare him an unbeliever after he has been informed, if he knows and [still] rejects."

> وأما الكذب والبهتان، فمثل قولهم: إنا نكفر بالعموم، ونوجب الهجرة إلينا على من قدر على إظهار دينه، وإنا نكفر من لم يكفر، ومن لم يقاتل ... وإذا كنا لا نكفر من عبد الصنم، الذي على عبد القادر، والصنم الذي على قبر أحمد البدوي، وأمثالهما، لأجل جهلهم، وعدم من ينبههم، فكيف نكفر من لم يشرك بالله إذا لم يهاجر إلينا، أو لم يكفر ويقاتل؟ ... بل نكفر تلك الأنواع الأربعة، لأجل محادتهم لله ورسوله

"As for the lies and slander, such as their saying that we declare unbelief wholesale, that we make emigration to us obligatory on anyone able to practice his religion openly, and that we declare unbelievers those who do not declare [others] unbelievers and do not fight: … If we do not declare unbelievers those who worship the idol on ['Abd al-Qadir's tomb] and the idol on Ahmad al-Badawi's grave and their like, because of their ignorance and the absence of anyone to alert them, how could we declare unbeliever someone who does not associate partners with God, merely because he did not emigrate to us or did not declare [others] unbelievers and fight? … Rather we declare those four types unbelievers because of their enmity to God and His Messenger."

The "four types" (pp. 102–104) are:
1. one who knows tawhid is the religion and that the beliefs about "stone, tree and man" are shirk, yet does not follow tawhid;
2. one who knows this but reviles the religion and praises those who worship Yusuf, al-Ashqar and others;
3. one who follows tawhid but hates those who entered it and loves those who remained in shirk;
4. one whose townspeople show enmity to "the people of tawhid" and who "fights the people of tawhid along with the people of his town … this one also is an unbeliever".

A textual tension to note plainly. Nawaqid no. 3 says that whoever does not declare the polytheists unbelievers is an unbeliever. The Durar letter calls it "slander" that "we declare unbelievers those who do not declare [others] unbelievers". The two may be reconciled by the condition "after being informed", but that reconciliation is an interpretation, not stated in either text.

### 3d. The Wahhabi attack on Karbala, 1216 AH

**Ibn Bishr, *'Unwan al-majd fi tarikh Najd*, ed. 'Abd al-Rahman b. 'Abd al-Latif Al al-Shaykh, 4th ed., Riyadh 1983, vol. 1, pp. 257–258**, "events of the year 1216". Transcribed by me from the page images, archive.org item `1-.-.-4`, PDF pages 258–259. [V-primary, page image]

> ثم دخلت السنة السادسة عشر بعد المائتين والألف، وفيها سار سعود بالجيوش المنصورة والخيل العتاق المشهورة من جميع حاضر نجد وباديها والجنوب والحجاز وتهامة وغير ذلك وقصدوا أرض كربلاء ونازل أهل بلد الحسين. وذلك في ذي القعدة فحشد عليها المسلمون، وتسوروا جدرانها ودخلوها عنوة، وقتلوا غالب أهلها في الأسواق والبيوت، وهدموا القبة الموضوعة بزعم من اعتقد فيها على قبر الحسين. وأخذوا ما في القبة وما حولها، وأخذوا النصيبة التي وضعوها على القبر، وكانت مرصوفة بالزمرد والياقوت والجواهر، وأخذوا جميع ما وجدوا في البلد من الأموال والسلاح واللباس والفرش والذهب والفضة والمصاحف الثمينة وغير ذلك مما يعجز عنه الحصر، ولم يلبثوا فيها إلا ضحوة وخرجوا منها قرب الظهر بجميع تلك الأموال وقتل من أهلها قريب ألفا رجل.
> ثم إن سعود ارتحل منها على الماء المعروف بالأبيض، فجمع الغنائم وعزل أخماسها وقسم باقيها على المسلمين غنيمة، للراجل سهم وللفارس سهمان. ثم ارتحل قافلا الى وطنه.

"Then the year 1216 began. In it Sa'ud marched with the victorious armies and the famous thoroughbred horses, from all the settled and nomadic people of Najd, the south, the Hijaz, Tihama and elsewhere. They headed for the land of Karbala and besieged the people of the town of al-Husayn. That was in Dhu al-Qa'da. The Muslims [i.e., Sa'ud's force] massed against it, scaled its walls and entered it by force. They killed most of its people in the markets and the houses, and demolished the dome that had been placed, according to the claim of those who venerated it, over the grave of al-Husayn. They took what was in the dome and around it, and took the covering-structure (al-nasiba) they had placed over the grave, which was inlaid with emerald, ruby and jewels. They took everything they found in the town, of wealth, weapons, clothing, furnishings, gold, silver, precious Qur'an copies and more than can be counted. They stayed there only through the forenoon and left near midday with all that wealth, and nearly two thousand men of its people were killed. Then Sa'ud moved to the water known as al-Abyad, gathered the spoils, set aside their fifths, and divided the rest among the Muslims as booty (ghanima): one share for the foot soldier and two for the horseman. Then he returned home."

Date: Dhu al-Qa'da 1216 runs from about 5 March to 3 April 1802 (tabular conversion). I did not verify the other dates for the attack given by other sources, including those placing it in Dhu al-Hijja 1216 / April 1802. [unverified]

Source character: Ibn Bishr (d. 1290/1873) was a Najdi chronicler sympathetic to the Saudi-Wahhabi state. His dates (1210–1290 AH / 1795–1873) appear to be his life dates in the title of a thesis about him listed on archive.org (item 1210129017951873p.d.f4968) [secondary]. If so, he was a child in 1216 and wrote decades later. He is the standard Najdi account, but he is not an adult eyewitness.

Legal point (my inference from the text): dividing the spoils as ghanima, with the khums set aside, applies the rules for war against unbelievers. Under Maliki and Shafi'i rules for Muslim rebels (bughat), their property is not taken as booty. Khalil: they are not enslaved and their property is returned; 'Umdat al-salik: "the rulings of Islam remain in force on them". [rules V-primary; the application to Karbala is my inference]

### 3e. Ibn Taymiyya on the Nusayris and Batinis (a Salafi-side parallel to al-Ghazali)

*Al-Fatawa al-kubra*, Kitab al-hudud, masa'il 755–757. Arabic Wikisource, page "الفتاوى الكبرى/كتاب الحدود/6". The same text is in *Majmu' al-fatawa* vol. 35, but I did not check that volume or its page. [V-primary, Wikisource]

> هؤلاء القوم المسمون بالنصيرية هم وسائر أصناف القرامطة الباطنية أكفر من اليهود والنصارى؟ بل وأكفر من كثير من المشركين وضررهم على أمة محمد ﷺ أعظم من ضرر الكفار المحاربين مثل كفار التتار والفرنج

"These people called the Nusayriyya, together with all kinds of Qarmati Batinis, are greater unbelievers than the Jews and Christians, indeed greater than many polytheists, and their harm to the community of Muhammad is greater than the harm of the warring unbelievers such as the Tatars and the Franks."

On a Nusayri group that "openly professed unbelief … left obedience and resolved on war" (masa'ala 757):

> هؤلاء يجب قتالهم ما داموا ممتنعين حتى يلتزموا شرائع الإسلام ... وهم مرتدون من أسوأ الناس ردة تقتل مقاتلهم وتغنم أموالهم

"These must be fought as long as they resist, until they commit to the laws of Islam … they are apostates of the worst kind; their fighters are killed and their property taken as booty."

On repentance: "there is disagreement among the scholars". Because "their doctrine's basis is taqiyya", he advises precaution: "they are not left gathered together, nor allowed to carry weapons" (فلا يتركون مجتمعين ولا يمكنون من حمل السلاح).

---

## Q4. Heresy, modern

### 4a. The Grozny conference, 25–27 August 2016: "Who are Ahl al-Sunna wa-l-Jama'a?"

**Final statement.** I did not locate it on an official conference or Chechen website. Quoted by Arabi21 (https://arabi21.com/story/943869) [secondary]:

> أهل السنة والجماعة هم الأشاعرة والماتريدية في الاعتقاد، وأهل المذاهب الأربعة الحنفية والمالكية والشافعية والحنابلة في الفقه، وأهل التصوف الصافي علما وأخلاقا وتزكية على مسلك الإمام الجنيد وأمثاله من أئمة الهدى

"Ahl al-Sunna wa-l-Jama'a are the Ash'aris and Maturidis in creed; the followers of the four schools, Hanafi, Maliki, Shafi'i and Hanbali, in law; and the people of pure Sufism in knowledge, ethics and purification, on the path of Imam al-Junayd and similar imams of guidance."

The statement does not list "ahl al-hadith" or Salafis as a creedal component.

- Habib 'Ali al-Jifri, a participant, argued that ahl al-hadith were included in substance, because they "were either consigners (mufawwid) of meaning or interpreters" of the divine attributes, "which is the creed of the Ash'aris and Maturidis", while "corporealists" (mujassima) were not. [secondary: quoting his Facebook post, via Arabi21]
- Arabi21 (story 943492) reports the attendees included Ahmed al-Tayeb, Ali Gomaa, Shawki Allam, al-Jifri, Sa'id Foudah and Hatim al-'Awni. The event was sponsored by Ramzan Kadyrov. [secondary]

**Was al-Azhar's Grand Imam there?** Yes. He gave a speech at the opening session [secondary: Gate Ahram 1231442; Al-Ain; Arabi21]. Al-Azhar and al-Tayeb then disowned the final statement. Al-Ain (https://al-ain.com/article/azhar-tayeb-grozny) quotes him [secondary]:

> لم نعلم عنه شيئًا، ولم يعرض علينا قبل إصداره، بل لم يكن وفد الأزهر الرسمي موجودًا في جروزني حين أعلن، ومن ثَمَّ فإن الأزهر غير مسؤول عن هذا البيان، وما يُسأل عنه فقط هو الكلمة التي ألقاها شيخ الأزهر

"We knew nothing of it; it was not shown to us before issue. The official Azhar delegation was not even present in Grozny when it was announced. Therefore al-Azhar is not responsible for this statement; it is answerable only for the speech the Shaykh al-Azhar gave."

He quoted his own speech:

> أنَّ أهل السُّنَّة والجَمَاعة هُم الأشاعرة والماتريدية، وأهل الحديث -وهم السلفيون- وأنَّ فُقَهَاء الحنفيَّة والمالكية والشافعية والحنابلة، لم يَخرُجوا من عَباءةِ هذا المذهب

"… that Ahl al-Sunna wa-l-Jama'a are the Ash'aris and Maturidis and ahl al-hadith, who are the Salafis, and that the jurists of the Hanafis, Malikis, Shafi'is and Hanbalis did not depart from the cloak of this school."

Asharq al-Awsat (https://aawsat.com/node/787826) quotes him: "O people of the Salaf, you are among Ahl al-Sunna wa-l-Jama'a" (يا أهل السلف أنتم من جملة أهل السنة والجماعة). [secondary]

**Saudi response.** The General Secretariat of the Council of Senior Scholars issued a statement that did not name Grozny explicitly. It said "the Muslim community is one community, and dividing it into parties and sects is a calamity the sharia did not bring" (أمة الإسلام أمة واحدة، وتفريقها إلى أحزاب وفرق من البلاء الذي لم تأت به الشريعة), and called for "fairness and justice in speech and judgment". [secondary: Al Jazeera, 2 Sep 2016; al-Watan, 10 Sep 2016]

Salafi reactions in Egypt (Yasir Burhami of al-Da'wa al-Salafiyya) accused the conference of "exclusion" of Salafis. [secondary: Arabi21 944075]

### 4b. Pakistan: Ahmadis

**Constitution (Second Amendment) Act, 1974.** President's assent 17 September 1974; Gazette of Pakistan, Extraordinary, Part I, 21 September 1974. Text from pakistani.org, an unofficial compilation (http://www.pakistani.org/pakistan/constitution/amendments/2amendment.html). [V-primary text, unofficial site]

- §2 inserts into Art. 106(3) the words "and persons of Quadiani group or the Lahori group (who call themselves 'Ahmadis')", among the minorities given reserved seats.
- §3 adds Art. 260(3):

> "A person who does not believe in the absolute and unqualified finality of The Prophethood of MUHAMMAD (Peace be upon him), the last of the Prophets or claims to be a Prophet, in any sense of the word or of any description whatsoever, after MUHAMMAD (Peace be upon him), or recognizes such a claimant as a Prophet or religious reformer, is not a Muslim for the purposes of the Constitution or law."

**Ordinance XX of 1984** ("Anti-Islamic Activities of Quadiani Group, Lahori Group and Ahmadis (Prohibition and Punishment) Ordinance, XX of 1984"). Per the footnote in the pakistani.org Pakistan Penal Code, it inserted PPC ss. 298-B and 298-C. Text of 298-C (http://www.pakistani.org/pakistan/legislation/1860/actXLVof1860.html) [V-primary text, unofficial site]:

> "Any person of the Quadiani group or the Lahori group (who call themselves 'Ahmadis' or by any other name), who directly or indirectly, poses himself as a Muslim, or calls, or refers to, his faith as Islam, or preaches or propagates his faith, or invites others to accept his faith, by words, either spoken or written, or by visible representations, or in any manner whatsoever outrages the religious feelings of Muslims shall be punished with imprisonment of either description for a term which may extend to three years and shall also be liable to fine."

Section 298-B penalizes Ahmadis' use of epithets reserved for the Prophet's family and Companions, and calling their place of worship a masjid, among other things.

**Which scholarly groups supported them.**
- Dawn (Shaikh Aziz, "A leaf from history: the Ahmadi issue", 23 Feb 2013) names the "Majlis Tahafuz-i-Khatme-Nabuwat, an alliance of 18 religio-political parties" as pressing Bhutto in 1974. [secondary]
- The Express Tribune (article 96375) says "In 1984, a group of religious leaders issued an ultimatum to Zia to take more measures against Ahmadis", which led to the ordinance. [secondary]
- A. Azfar Moin (*JRAS* 33.4, 2023, 1093–1106) describes the 1974 Special Committee proceedings as conducted by the attorney general "aided by the ulema members of parliament". [secondary, abstract only]
- **Not verified:** the specific claims that the resolution was moved by Shah Ahmad Noorani (Jamiat Ulama-e-Pakistan, Barelvi) and backed by Mufti Mahmud (Jamiat Ulama-e-Islam, Deobandi) and Jamaat-e-Islami; and the April 1974 Muslim World League (Rabita) resolution in Mecca declaring Ahmadis non-Muslim. I found these only in sources I could not open or should not cite. Check Ali Usman Qasmi, *The Ahmadis and the Politics of Religious Exclusion in Pakistan* (2014), and the National Assembly's published proceedings of the 1974 Special Committee. [unverified]

### 4c. Salafi and mainstream statements on Ash'aris and on Shi'a

**Ibn Uthaymeen, *Sharh al-'Aqida al-Wasitiyya*, ed. Sa'd al-Sumayl, Dar Ibn al-Jawzi, 6th ed. 1421, vol. 1, pp. 53–54** (Shamela 11250, page 51: https://shamela.ws/book/11250/51). [V-primary]

> وعلم من كلام المؤلف رحمه الله أنه لا يدخل فيهم من خالفهم في طريقتهم، فالأشاعرة مثلا والماتريدية لا يعدون من أهل السنة والجماعة في هذا الباب، لأنهم مخالفون لما كان عليه النبي صلى الله عليه وسلم وأصحابه في إجراء صفات الله سبحانه وتعالى على حقيقتها، ولهذا يخطئ من يقول: إن أهل السنة والجماعة ثلاثة: سلفيون، وأشعريون، وماتريديون، فهذا خطأ ... فأهل السنة والجماعة هم السلف معتقداً

"It is known from the author's words that those who differ from them in their method are not included among them. The Ash'aris, for example, and the Maturidis are not counted among Ahl al-Sunna wa-l-Jama'a *in this chapter*, because they differ from what the Prophet and his Companions held about taking God's attributes according to their reality. Hence it is a mistake to say that Ahl al-Sunna wa-l-Jama'a are three, Salafis, Ash'aris and Maturidis … Ahl al-Sunna wa-l-Jama'a are the Salaf in creed."

Note the qualifier "في هذا الباب", meaning in the chapter on divine attributes. This is a statement about orthodoxy, not takfir.

IslamQA no. 226290 (4 March 2015; https://islamqa.info/ar/answers/226290) [V-primary for IslamQA; secondary for the Ibn Uthaymeen quotations] gives the two-sense framework:
- **broad sense:** "Ahl al-Sunna" as opposed to the Rafida includes Ash'aris, Maturidis "and even the Mu'tazila";
- **strict sense:** as opposed to innovation, it includes only Salafis and ahl al-hadith.

It quotes Ibn Uthaymeen (*al-Sharh al-mumti'* 11/306) to the same effect. It also quotes him (*Liqa' al-bab al-maftuh* no. 43) as declining to call al-Nawawi and Ibn Hajar "Ash'aris" and praising their works.

**Ibn Baz, *Majmu' fatawa* 28/257** (https://binbaz.org.sa/fatwas/20154). [V-primary]

> الأشاعرة من أهل السنة في غالب الأمور، ولكنهم ليسوا منهم في تأويل الصفات، وليسوا بكفار بل فيهم الأئمة والعلماء والأخيار، ولكنهم غلطوا في تأويل بعض الصفات

"The Ash'aris are among Ahl al-Sunna in most matters, but not in their figurative interpretation of the attributes. They are not unbelievers; among them are imams, scholars and good people, but they erred in interpreting some attributes."

**Al-Fawzan on the Ash'aris:** not verified. I found only a forum reproduction of a lecture outline with a heading saying the Ash'aris and Maturidis are not Ahl al-Sunna. His signature on Permanent Committee fatwa 21166 is verified, but that fatwa is on apostasy, not on the Ash'aris. [unverified]

**Mainstream (Dar al-Ifta) on Ash'aris and on those who exclude them.** Fatwa no. 2370, Shawki 'Allam, dated 24 July 2013, "The ruling on accusing the Ash'aris of leaving Ahl al-Sunna wa-l-Jama'a" (https://www.dar-alifta.org/ar/fatwa/details/12581). [V-primary]

> ما رأيكم فيما يفعله بعض الوهابيين من رمي الأشاعرة بالخروج عن أهل السنة والجماعة؟ — مثل هذا الكلام لا يُعَوَّل عليه ولا يُلتَفت إليه؛ فلا يزال السادة الأشاعرة هم جمهور العلماء من الأمة ... ومَنْ كفَّرهم أو فسَّقهم أو شكَّك في عقيدتهم فإنه يُخْشَى عليه في دينه.

"[Q:] What is your view of what some Wahhabis do in accusing the Ash'aris of leaving Ahl al-Sunna wa-l-Jama'a? [A:] Such talk is not to be relied on or heeded; the Ash'aris remain the majority of the community's scholars … Whoever declares them unbelievers, or sinners, or casts doubt on their creed, there is fear for his religion."

The fatwa then quotes with approval Ibn Rushd al-Jadd (*Fatawa* 2/802) on those who revile the Ash'ari imams:

> فيجب أن يُبَصَّرَ الجاهلُ منهم، ويُؤَدَّبَ الفاسقُ، ويُسْتَتابَ المبتدعُ الزائغ عن الحق إذا كان مستسهلًا ببدعة، فإن تاب وإلا ضُرِب أبدًا حتى يتوب

"The ignorant among them must be enlightened, the sinner disciplined, and the innovator who deviates from the truth, if he makes light of his innovation, must be called to repent; if he repents [he is left], otherwise he is beaten continually until he repents."

A companion fatwa no. 16614 (same date) says the Ash'aris "are among Ahl al-Sunna, indeed they are the imams of Ahl al-Sunna and most of them". It quotes Ibn 'Abidin that Ahl al-Sunna "are the Ash'aris and Maturidis", and al-Saffarini that Ahl al-Sunna are three groups: Athariyya (led by Ahmad), Ash'ariyya and Maturidiyya. [V-primary]

**Salafi on Shi'a.**
- Ibn Baz (https://binbaz.org.sa/fatwas/4859) [V-primary] distinguishes types of Shi'a. Those who only rank 'Ali above 'Uthman or above Abu Bakr and 'Umar, without cursing or worshipping 'Ali, "their matter is easy; it does not harm them in religion" (أمرهم سهل، لا يضرُّهم ذلك من جهة الدين). Those who curse Mu'awiya or 'A'isha "erred and sinned". The "Batini" groups, "the Rafida, the Nusayris and their like", who "worship 'Ali … call on them for help … believe they know the unseen … curse the Companions and declare them unbelievers", are separated from Ahl al-Sunna "as heaven from earth".
- Ibn Baz (https://binbaz.org.sa/fatwas/30988, from an audio lecture) [V-primary]: "The Shi'a are among the most vile of the people of innovation, and the worst of them are the Imami Rafida … they have shirk and kufr … their unbelief is deliberate, like the Jews'" (الشيعة من أخبث أهل البدع، وشرهم الرافضة الإمامية ... لأن كفرهم متعمد مثل اليهود). The prescribed response in that answer is argument from the Qur'an and Sunna, not force.

**Mainstream on Shi'a.**
- Al-Ghazali (3a) does not judge an Imami commoner an unbeliever for the imamate doctrine, provided he does not hold Sunnis to be unbelievers.
- Ebussuud (3b) says the Kızılbaş are "not of the Shi'a" but a new mixture of unbeliefs, so his takfir is aimed at that group, not at Shi'ism as such.
- The **Amman Message** (2004–2005), whose takfir clause the site attributes to "the Shaykh al-Azhar's fatwa" (Muhammad Sayyid Tantawi), states: "Whosoever is an adherent to one of the four Sunni schools … the two Shi'i schools of Islamic jurisprudence (Ja'fari and Zaydi), the Ibadi school … and the Thahiri school … is a Muslim. Declaring that person an apostate is impossible and impermissible … it is neither possible nor permissible to declare whosoever subscribes to the Ash'ari creed or whoever practices real Tasawwuf (Sufism) an apostate. Likewise, it is neither possible nor permissible to declare whosoever subscribes to true Salafi thought an apostate." (https://ammanmessage.com/the-three-points-of-the-amman-message-v-1/) [V-primary, official site in English; Arabic text not checked]

---

## Q5. Blasphemy (brief)

**Pakistan Penal Code s. 295-C**, inserted by the Criminal Law (Amendment) Act III of 1986, s. 2, per the pakistani.org footnote. [V-primary text, unofficial site]

> "Use of derogatory remarks, etc., in respect of the Holy Prophet: Whoever by words, either spoken or written, or by visible representation or by any imputation, innuendo, or insinuation, directly or indirectly, defiles the sacred name of the Holy Prophet Muhammad (peace be upon him) shall be punished with death, or imprisonment for life, and shall also be liable to fine."

The Federal Shariat Court's 1990 ruling (*Muhammad Ismail Qureshi v. Pakistan*, reported as PLD 1991 FSC 10) is widely reported to have held the "imprisonment for life" alternative repugnant to Islam, leaving death as the only penalty. I did not read the judgment, and the pakistani.org text still shows the alternative. [unverified]

Who supports 295-C: not researched in primary sources in this pass. [unverified]

**Classical Maliki rule** (*Mukhtasar Khalil*, p. 239; Shamela 11355/308) [V-primary]:

> وإن سب نبيا أو ملكا أو عَرَّضَ أَوْ لَعَنَهُ أَوْ عَابَهُ أَوْ قَذَفَهُ أَوْ اسْتَخِفَّ بِحَقِّهِ ... قُتِلَ وَلَمْ يُسْتَتَبْ حَدًّا إلا أن يسلم الكافر

"If he reviles a prophet or an angel, or alludes [insultingly], curses him, finds fault with him, slanders him, or belittles his right … he is killed, and is not called to repent, as a hadd, unless [the reviler is] an unbeliever who becomes Muslim."

So, in the Maliki manual, blasphemy against a prophet is treated more severely than apostasy: no istitaba. The Permanent Committee's fatwa 21166 lists reviling God or the Messenger as apostasy "by word" (2a).

---

## Where mainstream Sunni and Salafi sources agree and differ

**Agree (historically):**
- Both traditions use the same hadith, the same consensus claim (*al-Mughni*: "the people of knowledge agree … that killing the apostate is obligatory"), and the same four-school manuals.
- Both require adulthood, sanity and free choice, and (in three schools) a call to repent before execution.
- Both treat secret or Batini unbelief as a special category where repentance may be refused: al-Ghazali, Khalil, al-Mughni, Ebussuud and Ibn Taymiyya all discuss this.
- Both assign execution to the ruler and treat private killing as a wrong. The penalty for that wrong is only ta'zir, with no blood-money (*Hidaya*: nothing due; *'Umda* and *Mughni*: ta'zir, no compensation; Ebussuud no. 486: no action against a private killer if the killing coincided with the act of cursing).
- Both have a two-tier scheme. "Innovation" (bid'a, dalal) is restrained but not killed. "Unbelief" is treated as apostasy. Al-Ghazali, Ibn Baz on the Ash'aris, and Ibn Baz on the "easy" Shi'a all use the lower tier.
- Both, in practice, applied rules of war against apostates to whole groups: the Batiniyya (al-Ghazali: "we assassinate them"), the Kızılbaş (Ebussuud: greater jihad, enslaving women), Karbala in 1802 (Ibn Bishr: most inhabitants killed, booty divided), the Nusayris (Ibn Taymiyya).

**Differ:**
- *What counts as unbelief.* Salafi texts (Nawaqid nos. 1–2; Kashf; Permanent Committee fatwa 21166) count calling on saints or the dead for intercession, and sacrificing at graves, as apostasy-level shirk. Ibn 'Abd al-Wahhab says it was this, not mere rebellion, that made "blood and property lawful". In the sources I read, the mainstream texts do not list these acts as nullifiers. I did not collect an Ash'ari or Sufi source specifically defending tawassul or istighatha in this pass, so the positive mainstream position on that point is not documented here.
- *Conditions on takfir.* Ibn 'Abd al-Wahhab requires "ta'rif" (being informed) and excuses the ignorant at shrines (*Durar* 1/104). Al-Ghazali requires examination of belief or upright witnesses. Ebussuud exempts peaceable villagers "whose outward state indicates their truthfulness". These are three different limiting procedures, but all three exist.
- *Today, on apostasy.* Egypt's official religious leaders (Gomaa 2007; al-Tayeb 2016) restate the punishment as conditional on harm to society. Al-Tayeb calls it ta'zir, which he says can range from reproach to death, and he classifies the public advocate of apostasy as committing "high treason". The Saudi Permanent Committee (fatwa 21166, signed by al-Fawzan among others) and Ibn Baz restate the classical death penalty for apostasy as such, with istitaba, and justify it by deterrence and protection of the community. The Committee also states (fatwa 4792) that one who has already repented is not punished.
- *Today, on who is "Sunni".* Grozny's final statement and Dar al-Ifta make Ash'ari/Maturidi creed the definition, though al-Tayeb personally included Salafis. Ibn Uthaymeen excludes the Ash'aris from Ahl al-Sunna "in this chapter". Ibn Baz says they are Ahl al-Sunna "in most matters" and "not unbelievers". The dispute is mainly about the label of orthodoxy, not about blood.
- *Lay opinion versus official doctrine.* In Egypt, whose official scholars now reinterpret the rule, Pew found the highest support for executing those who leave Islam (88% of all Muslims surveyed; 86% of sharia supporters). Pew did not survey Saudi Arabia, so no Saudi lay figure is available from this source.

---

## Bearing on Principles 4 and 5

1. **Known procedure.** The classical rule in all four manuals has a stated procedure:
   - adult, sane and acting by choice;
   - proof by testimony;
   - a call to repent, obligatory in three schools;
   - a fixed waiting period;
   - execution only by the ruler.
   
   This meets the "known procedure" part of Principle 5. The weak point is enforcement against private killers. The penalty for killing an apostate without authority is only discretionary punishment, with no blood-money or retaliation. That means the procedure protects the accused less than it protects other persons.

2. **Identical treatment of persons in the same position.** Within the category "Muslim who leaves Islam", the classical rule treats persons alike, with three exceptions:
   - the Hanafi rule treats women differently from men;
   - the Maliki rule, and some Hanbali and Hanafi narrations, refuse repentance from a "concealed" apostate, so two people who make the same statement of repentance can receive different outcomes depending on the category the judge assigns them;
   - Ebussuud leaves to the sultan which school's rule on repentance to apply.
   
   The rule is also asymmetric by design between entering and leaving: the Permanent Committee (fatwa 21166, q. 3) explicitly states that entering Islam is rewarded and leaving it is punished by death. Principle 5 does not forbid different rules for different positions, but here the "position" is defined by a change of belief.

3. **Enemy versus difference of opinion (Principle 4).** The sources divide on this:
   - The Hanafi rationale in *al-Hidaya* ties the death penalty to preventing warfare (hirab), and exempts women because warfare is not expected from them. That ties the penalty to expected action, not to belief alone.
   - Gomaa's 2007 text goes further: "If the case in question is one of merely rejecting faith, then there is no worldly punishment." Worldly penalty arises only when apostasy is combined with "undermining the foundations of the society" or "sedition". That matches Principle 4's distinction.
   - Al-Tayeb's 2016 remarks keep a ta'zir penalty that can reach death for the apostate who publicly advocates apostasy, which he calls "high treason". This counts speech as enemy action.
   - The Saudi Permanent Committee and Ibn Baz attach death to the change of religion itself, after a call to repent. Their stated reasons are deterrence and protecting the faith of others. Under Principle 4, that treats a person who holds a different opinion, and has not acted to rob, subjugate or destroy anyone, as punishable in the way an enemy would be.

4. **Group verdicts versus individual verdicts.** The historical cases decided matters by group category, not by individual trial:
   - al-Ghazali on the Batiniyya ("we assassinate them … not only when fighting");
   - Ebussuud on the Kızılbaş;
   - the Saudi raid on Karbala;
   - Ibn Taymiyya on the Nusayris.
   
   In each case there was a limiting clause: al-Ghazali's two grades and his acceptance of commoners' repentance; Ebussuud's exemption of peaceable townspeople and villagers "as long as their lying has not become evident"; Ibn 'Abd al-Wahhab's excuse for ignorance. Where the group was in armed conflict with the ruler (the Kızılbaş under the Safavids; Ibn Taymiyya's Nusayris who "resolved on war"), the ruling partly matches Principle 4's "enemies are real". Where the ruling reaches people who have not fought, such as al-Ghazali's assassination clause, women under the non-Hanafi rule, and the residents of Karbala killed "in the markets and the houses", it does not meet Principle 4. It also does not meet Principle 5, because the person was not judged individually by the stated procedure.

5. **Modern heresy disputes.** Most modern statements on Ash'aris versus Salafis are about who holds the title "Ahl al-Sunna", not about penalties:
   - Ibn Baz: the Ash'aris are "not unbelievers";
   - the Saudi Council of Senior Scholars after Grozny: called for unity and "fairness and justice in … judgment";
   - the Amman Message: forbids takfir of Ash'aris, Sufis and "true Salafi thought".
   
   These fit Principle 4. Two items in the sources point the other way, one from each camp:
   - Dar al-Ifta fatwa 2370 approvingly quotes Ibn Rushd that a person who reviles the Ash'ari imams should be "beaten continually until he repents";
   - Pakistan's Ordinance XX criminalizes Ahmadis calling themselves Muslims or calling their faith Islam.
   
   The Pakistani law is a known statute applied identically to every member of a named group, so it is procedurally uniform within its scope. But the offense is self-description and preaching, not robbery, subjugation or destruction, so under Principle 4 it treats holders of a different opinion as offenders.
