
# 1: $1>$2
# 0: $1=$2
# -1: $1<$2
function version_compare() {
  OLD_IFS="$IFS"
  IFS="."
  versionArr1=($version1)
  versionArr2=($version2)
  IFS="$OLD_IFS"
  versionComPareRes=0
  for ((ii = 0; ii < 3; ii++)); do
    if [ ${#versionArr1[ii]} \> ${#versionArr2[ii]} ]; then
      versionComPareRes=1
      break
      echo 11
    elif [ ${#versionArr1[ii]} \< ${#versionArr2[ii]} ]; then
      versionComPareRes=-1
      break
      echo 22
    else
      continue
    fi
  done
}