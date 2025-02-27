import useRequest from "@/hooks/useRequest/useRequest";
import systemApi, { InstanceType } from "@/services/systemSetting";
import useRealTimeTraffic from "@/models/dataanalysis/useRealTimeTraffic";
import useTemporaryQuery, {
  openNodeDataType,
} from "@/models/dataanalysis/useTemporaryQuery";
import useDataSourceManage from "@/models/dataanalysis/useDataSourceManage";
import { useState } from "react";
import useWorkflow from "@/models/dataanalysis/useWorkflow";
import useManageNodeAndFolder from "@/models/dataanalysis/useManageNodeAndFolder";
import temporaryQueryApi from "@/services/temporaryQuery";
import dataAnalysisApi from "@/services/dataAnalysis";
import realtimeApi from "@/services/realTimeTrafficFlow";
import useIntegratedConfigs from "@/models/dataanalysis/useIntegratedConfigs";
import dataSourceManageApi from "@/services/dataSourceManage";
import { formatMessage } from "@@/plugin-locale/localeExports";
import { message } from "antd";
import useWorkflowBoard from "@/models/dataanalysis/useWorkflowBoard";
import { FIRST_PAGE } from "@/config/config";
export interface versionHistoryListType {
  list: any[];
  total: number;
}

const DataAnalysis = () => {
  const [navKey, setNavKey] = useState<string>();
  const [instances, setInstances] = useState<InstanceType[]>([]);
  const [currentInstances, setCurrentInstances] = useState<number>();
  // 数据集成运行结果的id
  const [resultId, setResultId] = useState<number>(0);
  // 打开的文件节点id
  const [openNodeId, setOpenNodeId] = useState<number>();
  // 打开的文件节点父级id
  const [openNodeParentId, setOpenNodeParentId] = useState<number>();
  const [openNodeData, setOpenNodeData] = useState<openNodeDataType>();
  // 节点修改后的value
  const [folderContent, setFolderContent] = useState<string>("");

  // 版本历史list
  const [versionHistoryList, setVersionHistoryList] =
    useState<versionHistoryListType>({ list: [], total: 0 });
  // 版本历史的分页
  const [currentPagination, setCurrentPagination] = useState<API.Pagination>({
    current: FIRST_PAGE,
    pageSize: 10,
    total: 0,
  });

  // 右侧边栏运行结果弹窗
  const [visibleResults, setVisibleResults] = useState<boolean>(false);
  const [userList, setUserList] = useState<any[]>([]);

  // 右侧运行列表数据
  const [resultsList, setResultsList] = useState<any>({});
  const [visibleResultsItem, setVisibleResultsItem] = useState<boolean>(false);

  // 运行list的分页
  const [currentResultsPagination, setCurrentResultsPagination] =
    useState<API.Pagination>({
      current: FIRST_PAGE,
      pageSize: 10,
      total: 0,
    });

  const realTimeTraffic = useRealTimeTraffic();
  const temporaryQuery = useTemporaryQuery();
  const workflow = useWorkflow();
  const dataSourceManage = useDataSourceManage();
  const manageNode = useManageNodeAndFolder();
  const integratedConfigs = useIntegratedConfigs();
  const workflowBoard = useWorkflowBoard();

  const changeOpenNodeId = (id?: number) => {
    setOpenNodeId(id);
  };

  const changeOpenNodeParentId = (parentId: number) => {
    setOpenNodeParentId(parentId);
  };

  const changeOpenNodeData = (value: any) => {
    setOpenNodeData(value);
  };

  const changeFolderContent = (str: string) => {
    setFolderContent(str);
  };

  const onChangeNavKey = (key: string) => {
    setNavKey(key);
  };

  const changeResultId = (num: number) => {
    setResultId(num);
  };

  const onChangeCurrentInstances = (value?: number) => {
    setCurrentInstances(value);
  };

  const changeVersionHistoryList = (value: versionHistoryListType) => {
    setVersionHistoryList(value);
  };

  /**
   * api
   */

  const doGetInstance = useRequest(systemApi.getInstances, {
    loadingText: false,
  });

  const doGetDatabase = useRequest(realtimeApi.getDataBaseList, {
    loadingText: false,
  });

  const doGetTables = useRequest(realtimeApi.getTableList, {
    loadingText: false,
  });

  const doGetNodeInfo = useRequest(dataAnalysisApi.getNodeInfo, {
    loadingText: false,
  });

  // Node
  const doCreatedNode = useRequest(dataAnalysisApi.createdNode, {
    loadingText: false,
  });

  const doUpdateNode = useRequest(dataAnalysisApi.updateNode, {
    loadingText: false,
  });

  const doDeleteNode = useRequest(dataAnalysisApi.deleteNode, {
    loadingText: false,
  });

  const doLockNode = useRequest(temporaryQueryApi.lockNode, {
    loadingText: false,
  });

  const doUnLockNode = useRequest(temporaryQueryApi.unLockNode, {
    loadingText: false,
  });

  const doRunCodeNode = useRequest(temporaryQueryApi.runCodekNode, {
    loadingText: {
      loading: formatMessage({
        id: "bigdata.models.dataAnalysis.runLoadingText",
      }),
      done: formatMessage({
        id: "bigdata.models.dataAnalysis.runLoadingDoneText",
      }),
    },
  });

  const doGetSourceList = useRequest(dataSourceManageApi.getSourceList, {
    loadingText: false,
  });

  const doNodeHistories = useRequest(dataAnalysisApi.getNodeHistories, {
    loadingText: false,
  });

  const doNodeHistoriesInfo = useRequest(dataAnalysisApi.getNodeHistoriesInfo, {
    loadingText: false,
  });

  const doResultsList = useRequest(dataAnalysisApi.getResultsList, {
    loadingText: false,
  });

  const doModifyResults = useRequest(dataAnalysisApi.modifyResults, {
    loadingText: false,
  });

  const doResultsInfo = useRequest(dataAnalysisApi.getResultsInfo, {
    loadingText: false,
  });

  // 调度配置
  const doCreatCrontab = useRequest(dataAnalysisApi.creatCrontab, {
    loadingText: false,
  });

  const doGetCrontabInfo = useRequest(dataAnalysisApi.getCrontabInfo, {
    loadingText: false,
  });

  const doUpdateCrontab = useRequest(dataAnalysisApi.updateCrontab, {
    loadingText: false,
  });

  const doDeleteCrontab = useRequest(dataAnalysisApi.deleteCrontab, {
    loadingText: false,
  });

  const doGetUsers = useRequest(dataAnalysisApi.getUsers, {
    loadingText: false,
  });

  // 获取文件信息
  const onGetFolderList = (id: number) => {
    id &&
      doGetNodeInfo.run(id).then((res: any) => {
        if (res?.code === 0) {
          setOpenNodeData(res.data);
          changeFolderContent(res.data.content);
        }
      });
  };

  // 是否修改
  const isUpdateStateFun = () => {
    return folderContent !== openNodeData?.content;
  };

  /**文件夹标题*/

  // 锁定节点
  const handleLockFile = (nodeId: number) => {
    if (openNodeData?.lockAt == 0 && nodeId) {
      doLockNode.run(nodeId).then((res: any) => {
        if (res.code == 0) {
          onGetFolderList(nodeId);
        }
      });
    }
  };

  // 解锁节点
  const handleUnLockFile = (nodeId: number) => {
    if (isUpdateStateFun()) {
      message.warning(
        formatMessage({ id: "bigdata.models.dataAnalysis.unlockTips" })
      );
      return;
    }
    nodeId &&
      doUnLockNode.run(nodeId).then((res: any) => {
        if (res.code == 0) {
          onGetFolderList(nodeId);
        }
      });
  };

  const handleGrabLock = (file: any) => {
    manageNode.doMandatoryGetFileLock.run(file?.id).then((res: any) => {
      if (res.code != 0) return;
      message.success(
        formatMessage({ id: "bigdata.components.FileTitle.grabLockSuccessful" })
      );
      onGetFolderList(file?.id);
    });
  };

  // 保存编辑后的文件节点
  const handleSaveNode = () => {
    const data: any = {
      name: openNodeData?.name,
      content: folderContent,
      desc: openNodeData?.desc,
      folderId: openNodeParentId,
    };
    openNodeId &&
      doUpdateNode.run(openNodeId, data).then((res: any) => {
        if (res.code == 0) {
          message.success(
            formatMessage({ id: "log.index.manage.message.save.success" })
          );
          onGetFolderList(openNodeId);
        }
      });
  };

  // run
  const handleRunCode = async (nodeId: number, func?: any) => {
    nodeId &&
      doRunCodeNode.run(nodeId).then((res: any) => {
        if (res.code == 0) {
          func(nodeId);
        }
      });
  };

  // 获取用户责任人list
  const getUserList = () => {
    doGetUsers.run().then((res: any) => {
      if (res.code == 0) {
        setUserList(res.data);
      }
    });
  };

  return {
    instances,
    currentInstances,
    navKey,

    setInstances,
    onChangeCurrentInstances,
    onChangeNavKey,

    resultId,
    changeResultId,

    folderContent,
    changeFolderContent,

    openNodeData,
    changeOpenNodeData,

    openNodeId,
    changeOpenNodeId,

    versionHistoryList,
    changeVersionHistoryList,

    currentPagination,
    setCurrentPagination,

    visibleResultsItem,
    setVisibleResultsItem,

    currentResultsPagination,
    setCurrentResultsPagination,

    visibleResults,
    setVisibleResults,

    openNodeParentId,
    changeOpenNodeParentId,
    isUpdateStateFun,

    onGetFolderList,

    doGetInstance,
    doGetDatabase,
    doGetTables,
    doGetNodeInfo,
    doGetSourceList,

    // node
    doCreatedNode,
    doUpdateNode,
    doDeleteNode,
    doLockNode,
    doUnLockNode,
    doRunCodeNode,

    // sqlTitle
    handleLockFile,
    handleUnLockFile,
    handleSaveNode,
    handleRunCode,
    handleGrabLock,

    // histories
    doNodeHistories,
    doNodeHistoriesInfo,

    // results
    doResultsList,
    doResultsInfo,
    doModifyResults,
    resultsList,
    setResultsList,

    // crontab
    doCreatCrontab,
    doGetCrontabInfo,
    doUpdateCrontab,
    doDeleteCrontab,
    userList,
    getUserList,

    manageNode,
    integratedConfigs,
    workflowBoard,
    realTimeTraffic,
    temporaryQuery,
    workflow,
    dataSourceManage,
  };
};

export default DataAnalysis;
