# 1: $1>$2
# 0: $1=$2
# -1: $1<$2
function a_b() {
  a=$1
  b=$2
  abRes=1
  if [ ${#a} \> ${#b} ]; then
    return
  elif [ ${#a} \< ${#b} ]; then
    abRes=-1
    return
  else
    if [ $a \> $b ]; then
      return
    elif [ $a \< $b ]; then
      abRes=-1
      return
    else
      abRes=0
      return
    fi
  fi

}
# 1: $1>$2
# 0: $1=$2
# -1: $1<$2
function version_compare() {
  versionComPareRes=1
  versionArr1_1=$(echo $version1 | awk -F ":" '{print $1}')
  versionArr1_2=$(echo $version1 | awk -F ":" '{print $2}')
  versionArr1_3=$(echo $version1 | awk -F ":" '{print $3}')

  versionArr2_1=$(echo $version2 | awk -F ":" '{print $1}')
  versionArr2_2=$(echo $version2 | awk -F ":" '{print $2}')
  versionArr2_3=$(echo $version2 | awk -F ":" '{print $3}')

  a_b $versionArr1_1 $versionArr2_1
  if [ $abRes -gt 0 ]; then
    return
  elif [ $abRes -lt 0 ]; then
    versionComPareRes=-1
    return
  else

    a_b $versionArr1_2 $versionArr2_2
    if [ $abRes -gt 0 ]; then
      return
    elif [ $abRes -lt 0 ]; then
      versionComPareRes=-1
      return
    else

      a_b $versionArr1_3 $versionArr2_3
      if [ $abRes -gt 0 ]; then
        return
      elif [ $abRes -lt 0 ]; then
        versionComPareRes=-1
        return
      else
        versionComPareRes=0
        return
      fi
    fi
  fi
}
