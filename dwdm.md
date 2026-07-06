
ITU rec G.694.1
==========================================

From
G.694.1 2020-10-29 version 3
https://www.itu.int/itu-t/recommendations/rec.aspx?rec=14498

For C-Band, $`DWDMCenterFrequency (Hz) = 193.1 \times 10^{12}`$


# STANDARD GRIDS

## 12.5Ghz
For channel spacings of 12.5 GHz on a fiber, the allowed channel frequencies (in THz) are defined by:
$`193.1 + n × 0.0125`$ where n is a positive or negative integer including 0

## 25.0Ghz
For channel spacings of 25 GHz on a fiber, the allowed channel frequencies (in THz) are defined by:
$`193.1 + n × 0.025`$ where n is a positive or negative integer including 0

## 50.0Ghz
For channel spacings of 50 GHz on a fiber, the allowed channel frequencies (in THz) are defined by:
$`193.1 + n × 0.05`$ where n is a positive or negative integer including 0

## 100.0Ghz
For channel spacings of 100 GHz or more on a fiber, the allowed channel frequencies (in THz) are defined by:
$`193.1 + n × 0.1`$ where n is a positive or negative integer including 0

# NON-ITU GRIDS

Cf OIForum CMIS 5.4 ver.

## 75.0Ghz
tuning on the 75 GHz grid defined as:
$`Frequency (THz) = 193.1 + n × 0.025`$
where n must be an integer multiple of 3.

## 33.0Ghz
tuning on the 33 GHz grid defined as:
$`Frequency (THz) = 193.1 + n × 0.1/3`$

## 6.25Ghz
tuning on the 6.25 GHz grid defined as:
$`Frequency (THz) = 193.1 + n × 0.00625`$

## 3.125Ghz
tuning on 3.125 GHz grid defined as:
$`Frequency (THz) = 193.1 + n × 0.003125`$

## 150Ghz
tuning on the 150 GHz grid defined as:
$`Frequency (THz) = 193.1 + (n+3) × 0.025`$
where n must be an integer multiple of 6.

## 300Ghz
tuning on the 300 GHz grid defined as:
$`Frequency (THz) = 193.1 + (n-9) × 0.0125`$
where n must be an integer multiple of 24.


# CHANNEL OFFSET NUMBER

This channel offset number effectively specifies a laser frequency in terms of a (signed) frequency offset from
a nominal reference frequency at 193.1THz, in units of the grid resolution (except for the 75GHz, 150GHz
and 300GHz grids, where the offset is defined in units of a third, a sixth, or a 24th of the grid resolution,
respectively).

